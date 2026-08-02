# parsers/pdf_parser.py - PDF解析器核心类（清理版）

import time
import fitz  # PyMuPDF
import sys
import os
import io
import contextlib
from typing import Dict, Any, List
from pathlib import Path

# 添加父目录到路径，以便导入utils模块
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from utils.logger import get_logger

from .table_extractor_pdfplumber import HybridTableExtractor
from .image_extractor import ImageExtractor

logger = get_logger()

# 当文本块与表格区域的重叠比例超过该阈值时，认为属于表格内部
TEXT_TABLE_OVERLAP_THRESHOLD = 0.6
# 对于文本块中心落在表格内部的情况，允许更低的重叠比例即可认定
TEXT_TABLE_CENTER_OVERLAP_THRESHOLD = 0.2

@contextlib.contextmanager
def suppress_stderr():
    """临时抑制stderr输出，防止PyMuPDF错误污染stdout"""
    with open(os.devnull, "w") as devnull:
        old_stderr = sys.stderr
        sys.stderr = devnull
        try:
            yield
        finally:
            sys.stderr = old_stderr

# 尝试导入pdfplumber，如果失败则设置为None
try:
    import pdfplumber
    PDFPLUMBER_AVAILABLE = True
except ImportError:
    pdfplumber = None
    PDFPLUMBER_AVAILABLE = False
    logger.warning("pdfplumber not available, table extraction will use PyMuPDF only")


class PDFParser:
    """PDF解析器主类（清理版）"""

    def __init__(self, pdf_path: str, need_character: bool = False,
                 extract_images: bool = False, page_range: str = "all",
                 table_engine: str = "hybrid", high_accuracy: bool = False):
        """
        初始化PDF解析器

        Args:
            pdf_path: PDF文件路径
            need_character: 是否需要字符级信息
            extract_images: 是否提取图片
            page_range: 页面范围，如 "1-5", "1,3,5", "all"
            table_engine: 表格提取引擎，"pdfplumber", "pymupdf", "hybrid"
            high_accuracy: 是否启用高精度模式（启用Camelot兜底）
        """
        self.pdf_path = pdf_path
        self.need_character = need_character
        self.extract_images = extract_images
        self.page_range = page_range
        self.table_engine = table_engine
        self.high_accuracy = high_accuracy

        # 根据table_engine确定是否使用pdfplumber
        self.use_pdfplumber = (table_engine in ["pdfplumber", "hybrid"]) and PDFPLUMBER_AVAILABLE

        self.doc = None
        self.pdf_plumber_doc = None

        logger.info(f"Initializing PDF parser for: {pdf_path}")
        logger.debug(f"Parameters: need_character={need_character}, extract_images={extract_images}, "
                    f"page_range={page_range}, table_engine={table_engine}, "
                    f"use_pdfplumber={self.use_pdfplumber}")

    def parse(self) -> Dict[str, Any]:
        """解析PDF文档"""
        start_time = time.time()
        logger.info("Starting PDF parsing")

        try:
            # 打开PyMuPDF文档，抑制可能的错误输出
            try:
                with suppress_stderr():
                    self.doc = fitz.open(self.pdf_path)
                logger.info(f"Opened PDF with PyMuPDF: {len(self.doc)} pages")
            except Exception as e:
                error_msg = str(e)
                logger.error(f"Failed to open PDF with PyMuPDF: {error_msg}")

                # 检查是否是加密相关错误
                if "aes filter" in error_msg.lower() or "encryption" in error_msg.lower():
                    raise Exception(f"PDF file is encrypted or corrupted: {error_msg}")
                elif "premature end" in error_msg.lower():
                    raise Exception(f"PDF file is corrupted or truncated: {error_msg}")
                else:
                    raise Exception(f"Cannot open PDF file: {error_msg}")

            # 如果需要使用pdfplumber，提前打开文档
            if self.use_pdfplumber:
                try:
                    self.pdf_plumber_doc = pdfplumber.open(self.pdf_path)
                    logger.debug("Opened PDF with pdfplumber for table extraction")
                except Exception as e:
                    logger.warning(f"Failed to open PDF with pdfplumber: {e}")
                    self.pdf_plumber_doc = None
                    self.use_pdfplumber = False

            # 解析页面范围
            page_indices = self._parse_page_range(self.page_range, len(self.doc))
            logger.debug(f"Processing pages: {page_indices}")

            pages = []
            for page_num in page_indices:
                logger.debug(f"Processing page {page_num + 1}")
                try:
                    with suppress_stderr():
                        page = self.doc[page_num]
                        page_data = self._parse_page(page, page_num)
                    pages.append(page_data)
                except Exception as e:
                    logger.warning(f"Error processing page {page_num + 1}: {e}")
                    # 创建一个空页面数据，避免中断整个解析过程
                    page_data = {
                        "angle": 0,
                        "height": 0,
                        "width": 0,
                        "tables": []
                    }
                    pages.append(page_data)

            duration = int((time.time() - start_time) * 1000)  # 毫秒
            logger.info(f"PDF parsing completed in {duration}ms")

            result = {
                "message": "success",
                "code": 0,
                "version": "1.0.0",
                "duration": duration,
                "result": {
                    "pages": pages
                }
            }

            return result

        except Exception as e:
            duration = int((time.time() - start_time) * 1000)
            logger.exception(f"Error parsing PDF: {str(e)}")
            return self._create_error_response(str(e), duration)
        finally:
            # 安全关闭文档资源
            self._cleanup_resources()

    def _parse_page_range(self, page_range_str, total_pages):
        """解析页面范围字符串"""
        if page_range_str.lower() == "all":
            return list(range(total_pages))

        pages = set()

        for part in page_range_str.split(','):
            part = part.strip()

            if '-' in part:
                start, end = part.split('-')
                start = int(start) - 1
                end = min(int(end), total_pages)
                pages.update(range(start, end))
            else:
                page = int(part) - 1
                if 0 <= page < total_pages:
                    pages.add(page)

        return sorted(list(pages))

    def _parse_page(self, page, page_num: int) -> Dict[str, Any]:
        """解析单个页面"""
        try:
            width = int(page.rect.width)
            height = int(page.rect.height)
            logger.debug(f"Parsing page {page_num + 1} ({width}x{height})")
        except Exception as e:
            logger.warning(f"Error getting page dimensions for page {page_num + 1}: {e}")
            width = 595  # A4默认宽度
            height = 842  # A4默认高度

        # 页面基本信息
        page_data = {
            "angle": 0,
            "height": height,
            "width": width,
            "tables": []
        }

        # 添加图片信息（如果需要）
        if self.extract_images:
            try:
                # 传递PDF路径给图片提取器以支持外部工具
                image_extractor = ImageExtractor(self.doc, self.pdf_path)
                images = image_extractor.extract_images(page_num)
                page_data["images"] = images
                logger.debug(f"Extracted {len(images)} images from page {page_num + 1}")

                # 记录提取方法统计
                methods_used = {}
                for img in images:
                    method = img.get("extraction_method", "unknown")
                    methods_used[method] = methods_used.get(method, 0) + 1

                if methods_used:
                    logger.debug(f"Extraction methods used: {methods_used}")

            except Exception as e:
                logger.warning(f"Failed to extract images from page {page_num + 1}: {e}")
                page_data["images"] = []

        # 使用表格提取器
        logger.debug(f"Using table engine: {self.table_engine}")

        if self.table_engine == "pymupdf":
            # 只使用PyMuPDF
            tables = self._extract_tables_pymupdf_only(page, page_num)
        else:
            # 使用混合提取器（包括hybrid和pdfplumber模式）
            extractor = HybridTableExtractor(
                pdf_path=self.pdf_path,
                page=page,
                need_character=self.need_character,
                pdf_plumber_doc=self.pdf_plumber_doc,
                engine_mode=self.table_engine,
                high_accuracy=self.high_accuracy
            )
            tables = extractor.find_tables()

        logger.debug(f"Found {len(tables)} tables on page {page_num + 1}")

        # debug
        for i, table in enumerate(tables):
            logger.debug(f"Table {i}: position={table['position']}, type={table.get('type')}, "
                f"rows={table.get('table_rows')}, cols={table.get('table_cols')}")

        # 获取表格区域的边界，用于区分文本区域
        table_regions = []
        for table in tables:
            pos = table["position"]
            table_regions.append({
                'rect': fitz.Rect(pos[0], pos[1], pos[2], pos[3]),
                'table': table
            })

        # 提取表格外的文本块
        text_blocks = self._extract_text_blocks_outside_tables(page, table_regions)
        logger.debug(f"Found {len(text_blocks)} text blocks outside tables on page {page_num + 1}")

        # 将文本块和表格按照位置排序，构建统一的tables列表
        all_elements = self._combine_tables_and_text(tables, text_blocks)

        # 按照位置排序（从上到下，从左到右）
        all_elements.sort(key=lambda x: (x['position'][1], x['position'][0]))

        # 构建最终的tables列表
        page_data['tables'] = [elem['table'] for elem in all_elements]

        logger.debug(f"Page {page_num + 1} processed: {len(page_data['tables'])} elements total")

        return page_data

    def _extract_tables_pymupdf_only(self, page, page_num: int) -> List[Dict[str, Any]]:
        """只使用PyMuPDF提取表格的简化方法"""
        logger.debug("Using PyMuPDF-only table extraction")
        # 这里可以实现一个简化的PyMuPDF表格提取逻辑
        # 暂时返回空列表，实际使用时需要实现
        return []

    def _extract_text_blocks_outside_tables(self, page, table_regions):
        """提取表格外的文本块"""
        try:
            with suppress_stderr():
                text_dict = page.get_text("dict")
        except Exception as e:
            logger.warning(f"Error extracting text from page: {e}")
            return []

        text_blocks = []

        for block in text_dict['blocks']:
            if block['type'] == 0:  # 文本块
                try:
                    block_rect = fitz.Rect(block['bbox'])

                    # 收集非表格区域的文本块
                    for line in block['lines']:
                        try:
                            filtered_spans = []
                            for span in line['spans']:
                                text = span.get('text', '')
                                if not text or not text.strip():
                                    continue
                                bbox = span.get('bbox')
                                if not bbox:
                                    continue
                                if self._span_belongs_to_table(bbox, table_regions):
                                    continue
                                filtered_spans.append(span)

                            if not filtered_spans:
                                continue

                            line_text = ''.join(span['text'] for span in filtered_spans).strip()
                            if not line_text:
                                continue

                            # 重新计算文本块的bbox
                            x0 = min(span['bbox'][0] for span in filtered_spans)
                            y0 = min(span['bbox'][1] for span in filtered_spans)
                            x1 = max(span['bbox'][2] for span in filtered_spans)
                            y1 = max(span['bbox'][3] for span in filtered_spans)

                            text_blocks.append({
                                'bbox': [x0, y0, x1, y1],
                                'text': line_text,
                                'block_bbox': block['bbox']
                            })
                        except Exception as e:
                            logger.warning(f"Error processing text line: {e}")
                            continue
                except Exception as e:
                    logger.warning(f"Error processing text block: {e}")
                    continue

        return text_blocks

    def _span_belongs_to_table(self, span_bbox, table_regions):
        """判断文本span是否属于表格内部"""
        try:
            span_rect = fitz.Rect(span_bbox)
        except Exception:
            return False

        span_area = max(span_rect.get_area(), 1e-6)
        span_center_x = (span_rect.x0 + span_rect.x1) / 2

        for table_region in table_regions:
            table_rect = table_region['rect']
            inter = span_rect & table_rect
            inter_area = max(inter.get_area(), 0.0)
            if inter_area <= 0.0:
                continue

            overlap = inter_area / span_area
            if overlap >= TEXT_TABLE_OVERLAP_THRESHOLD:
                return True

            # 如果中心点落在表格内部，也视为属于表格
            if table_rect.x0 <= span_center_x <= table_rect.x1 and overlap >= TEXT_TABLE_CENTER_OVERLAP_THRESHOLD:
                return True

        return False

    def _combine_tables_and_text(self, tables, text_blocks):
        """合并表格和文本块"""
        all_elements = []

        # 添加文本块作为plain类型的table
        if text_blocks:
            plain_tables = self._group_text_blocks_into_plain_tables(text_blocks)
            for plain_table in plain_tables:
                all_elements.append({
                    'type': 'plain',
                    'position': plain_table['position'],
                    'table': plain_table
                })

        # 添加真正的表格
        for table in tables:
            # 规范化表格类型
            table_type = table.get('type', 'table')
            if table_type == 'table':
                table['type'] = 'table_with_line'
            elif table_type == 'table_no_border':
                table['type'] = 'table_without_line'

            all_elements.append({
                'type': table['type'],
                'position': table['position'],
                'table': table
            })

        return all_elements

    def _group_text_blocks_into_plain_tables(self, text_blocks):
        """将文本块分组成plain类型的table"""
        if not text_blocks:
            return []

        # 简化的文本块分组逻辑
        # 按垂直位置分组
        text_blocks_with_idx = [(i, block) for i, block in enumerate(text_blocks)]
        text_blocks_with_idx.sort(key=lambda x: (x[1]['bbox'][1], x[1]['bbox'][0]))

        plain_tables = []
        current_group = []
        vertical_gap_threshold = 20

        for idx, block in text_blocks_with_idx:
            if not current_group:
                current_group = [(idx, block)]
            else:
                last_block = current_group[-1][1]
                current_block_y = block['bbox'][1]
                last_block_y = last_block['bbox'][3]

                if current_block_y - last_block_y <= vertical_gap_threshold:
                    current_group.append((idx, block))
                else:
                    # 创建plain table
                    if current_group:
                        current_group.sort(key=lambda x: x[0])  # 按原始索引排序
                        blocks_only = [item[1] for item in current_group]
                        plain_table = self._create_plain_table(blocks_only)
                        plain_tables.append(plain_table)

                    # 开始新组
                    current_group = [(idx, block)]

        # 处理最后一组
        if current_group:
            current_group.sort(key=lambda x: x[0])
            blocks_only = [item[1] for item in current_group]
            plain_table = self._create_plain_table(blocks_only)
            plain_tables.append(plain_table)

        return plain_tables

    def _create_plain_table(self, text_blocks):
        """从文本块创建plain类型的table"""
        # 计算边界
        min_x = min(block['bbox'][0] for block in text_blocks)
        min_y = min(block['bbox'][1] for block in text_blocks)
        max_x = max(block['bbox'][2] for block in text_blocks)
        max_y = max(block['bbox'][3] for block in text_blocks)

        # 创建lines
        lines = []
        for block in text_blocks:
            tline = {
                "angle": 0,
                "text": block['text'],
                "direction": 0,
                "handwritten": 0,
                "position": [int(block['bbox'][0]), int(block['bbox'][1]),
                           int(block['bbox'][2]), int(block['bbox'][3])],
                "score": 1.0,
                "type": "text"
            }

            # 如果需要字符信息，这里可以进一步处理
            if self.need_character:
                char_info = self._init_char_info()
                tline.update(char_info)

            lines.append(tline)

        # 创建plain table
        plain_table = {
            "height_of_rows": [],  # plain类型不需要行高
            "type": "plain",
            "table_cells": [],  # plain类型没有单元格
            "table_rows": 0,
            "width_of_cols": [],  # plain类型不需要列宽
            "position": [int(min_x), int(min_y), int(max_x), int(max_y)],
            "lines": lines,
            "table_cols": 0
        }

        return plain_table

    def _init_char_info(self) -> Dict[str, List]:
        """初始化字符级信息结构"""
        return {
            "char_attributes": [],
            "char_candidates": [],
            "char_candidates_score": [],
            "char_centers": [],
            "char_positions": [],
            "char_scores": []
        }

    def _cleanup_resources(self):
        """安全清理资源"""
        # 清理pdfplumber文档
        if self.pdf_plumber_doc:
            try:
                self.pdf_plumber_doc.close()
                logger.debug("Closed pdfplumber document")
            except Exception as e:
                logger.warning(f"Error closing pdfplumber document: {e}")
            finally:
                self.pdf_plumber_doc = None

        # 清理PyMuPDF文档
        if self.doc:
            try:
                self.doc.close()
                logger.debug("Closed PyMuPDF document")
            except Exception as e:
                logger.warning(f"Error closing PyMuPDF document: {e}")
            finally:
                self.doc = None

    def _create_error_response(self, error_message: str, duration: int = 0) -> Dict[str, Any]:
        """创建标准错误响应"""
        return {
            "message": f"error: {error_message}",
            "code": -1,
            "version": "1.0.0",
            "duration": duration,
            "result": {"pages": []}
        }
