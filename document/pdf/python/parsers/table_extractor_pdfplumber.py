# parsers/table_extractor_pdfplumber.py - 基于pdfplumber的表格提取器（优化版）

import pdfplumber
import sys
import os
from typing import List, Dict, Any, Optional

# 添加父目录到路径，以便导入utils模块
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from utils.logger import get_logger

logger = get_logger()

try:
    import camelot
    CAMELOT_AVAILABLE = True
except Exception:
    camelot = None
    CAMELOT_AVAILABLE = False


class PDFPlumberTableExtractor:
    """基于pdfplumber的表格提取器 - 专门处理线框表格（优化版）"""
    
    def __init__(self, pdf_path: str, need_character: bool = False, pdf_plumber_doc=None):
        """
        初始化提取器
        
        Args:
            pdf_path: PDF文件路径
            need_character: 是否需要字符级信息
            pdf_plumber_doc: 预先打开的pdfplumber文档对象，如果提供则复用
        """
        self.pdf_path = pdf_path
        self.need_character = need_character
        self.pdf_plumber_doc = pdf_plumber_doc
        self._owns_doc = pdf_plumber_doc is None  # 标记是否由此类负责关闭文档
        
        # 如果没有提供文档对象，则打开文件
        if self.pdf_plumber_doc is None:
            self.pdf_plumber_doc = pdfplumber.open(self.pdf_path)
            logger.debug(f"Opened PDF with pdfplumber: {self.pdf_path}")
        else:
            logger.debug("Reusing existing pdfplumber document")
    
    def __enter__(self):
        """上下文管理器入口"""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """上下文管理器出口"""
        self.close()
    
    def close(self):
        """关闭PDF文档"""
        if self._owns_doc and self.pdf_plumber_doc:
            self.pdf_plumber_doc.close()
            self.pdf_plumber_doc = None
            logger.debug("Closed pdfplumber document")
        
    def extract_tables_from_page(self, page_num: int) -> List[Dict[str, Any]]:
        """从指定页面提取表格"""
        tables = []
        
        try:
            if page_num >= len(self.pdf_plumber_doc.pages):
                logger.warning(f"Page {page_num + 1} does not exist")
                return tables
            
            page = self.pdf_plumber_doc.pages[page_num]
            logger.debug(f"Processing page {page_num + 1} with pdfplumber")
            
            # 使用pdfplumber查找表格
            plumber_tables = page.find_tables(table_settings={
                "vertical_strategy": "lines",      # 基于线条的垂直策略
                "horizontal_strategy": "lines",    # 基于线条的水平策略
                "min_words_vertical": 1,          # 最少垂直单词数
                "min_words_horizontal": 1,        # 最少水平单词数
                "intersection_tolerance": 3,      # 交叉容差
                "join_tolerance": 3,              # 连接容差
                "edge_min_length": 20,            # 边的最小长度
                "snap_tolerance": 3,              # 对齐容差
            })
            
            logger.debug(f"pdfplumber found {len(plumber_tables)} tables on page {page_num + 1}")
            
            for table_idx, plumber_table in enumerate(plumber_tables):
                try:
                    # 提取表格数据
                    table_data = plumber_table.extract()
                    
                    if not table_data or len(table_data) == 0:
                        logger.debug(f"Table {table_idx + 1} has no data")
                        continue
                    
                    # 转换为我们的格式
                    converted_table = self._convert_plumber_table_to_our_format(
                        table_data, plumber_table.bbox, page_num, table_idx
                    )
                    
                    if converted_table:
                        tables.append(converted_table)
                        logger.debug(f"Successfully converted table {table_idx + 1}: {len(table_data)}x{len(table_data[0]) if table_data else 0}")
                    
                except Exception as e:
                    logger.warning(f"Failed to process table {table_idx + 1} on page {page_num + 1}: {e}")
                    continue
            
        except Exception as e:
            logger.error(f"Failed to process page {page_num + 1} with pdfplumber: {e}")
        
        return tables
    
    def _convert_plumber_table_to_our_format(self, table_data: List[List], bbox: tuple, 
                                           page_num: int, table_idx: int) -> Optional[Dict[str, Any]]:
        """将pdfplumber的表格数据转换为我们的格式"""
        
        if not table_data:
            return None
        
        # 过滤掉全空的行和列
        filtered_data = []
        for row in table_data:
            if any(cell and str(cell).strip() for cell in row if cell is not None):
                filtered_data.append(row)
        
        if not filtered_data:
            return None
        
        # 确保所有行的列数一致
        max_cols = max(len(row) for row in filtered_data)
        normalized_data = []
        for row in filtered_data:
            normalized_row = row[:max_cols]  # 截取到最大列数
            # 补齐缺失的列
            while len(normalized_row) < max_cols:
                normalized_row.append(None)
            normalized_data.append(normalized_row)
        
        rows = len(normalized_data)
        cols = max_cols
        
        if rows < 1 or cols < 1:
            return None
        
        # 计算单元格位置（简化计算）
        x0, y0, x1, y1 = bbox
        cell_width = (x1 - x0) / cols
        cell_height = (y1 - y0) / rows
        
        # 检测合并单元格
        merged_cells_data = self._detect_merged_cells(normalized_data)
        
        # 创建单元格和行数据
        table_cells = []
        table_lines = []  # 注意：表格中的lines应该为空，因为所有文本都在cells中
        processed_cells = set()  # 记录已处理的单元格位置
        
        for row_idx, row in enumerate(normalized_data):
            for col_idx, cell_value in enumerate(row):
                # 跳过已经被合并单元格处理的位置
                if (row_idx, col_idx) in processed_cells:
                    continue
                
                # 检查是否为合并单元格的起始位置
                merge_info = merged_cells_data.get((row_idx, col_idx))
                if merge_info:
                    start_row, start_col = row_idx, col_idx
                    end_row, end_col = merge_info['end_row'], merge_info['end_col']
                    merged_text = merge_info['text']
                    
                    # 标记所有被合并的单元格位置为已处理
                    for r in range(start_row, end_row + 1):
                        for c in range(start_col, end_col + 1):
                            processed_cells.add((r, c))
                else:
                    # 普通单元格
                    start_row = end_row = row_idx
                    start_col = end_col = col_idx
                    merged_text = str(cell_value).strip() if cell_value is not None else ""
                    processed_cells.add((row_idx, col_idx))
                
                # 计算合并单元格的位置
                cell_x0 = x0 + start_col * cell_width
                cell_y0 = y0 + start_row * cell_height
                cell_x1 = x0 + (end_col + 1) * cell_width
                cell_y1 = y0 + (end_row + 1) * cell_height
                
                # 创建TLine对象 - 只用于单元格内部，不添加到table_lines
                if merged_text:
                    tline = {
                        "angle": 0,
                        "text": merged_text,
                        "direction": 0,
                        "handwritten": 0,
                        "position": [int(cell_x0 + 2), int(cell_y0 + 2), 
                                   int(cell_x1 - 2), int(cell_y1 - 2)],
                        "score": 1.0,
                        "type": "text"
                    }
                    
                    # 如果需要字符信息
                    if self.need_character:
                        char_info = self._create_char_info_for_text(merged_text, tline["position"])
                        tline.update(char_info)
                    
                    cell_lines = [tline]
                else:
                    cell_lines = []
                
                # 创建单元格
                cell = {
                    "start_row": start_row,
                    "start_col": start_col,
                    "end_row": end_row,
                    "end_col": end_col,
                    "text": merged_text,
                    "borders": {
                        "right": 1,
                        "bottom": 1,
                        "left": 1,
                        "top": 1
                    },
                    "position": [int(cell_x0), int(cell_y0), int(cell_x1), int(cell_y1)],
                    "lines": cell_lines
                }
                table_cells.append(cell)
        
        # 计算行高和列宽
        height_of_rows = [int(cell_height)] * rows
        width_of_cols = [int(cell_width)] * cols
        
        # 创建表格对象
        table = {
            "height_of_rows": height_of_rows,
            "type": "table_with_line",  # pdfplumber主要检测有线表格
            "table_cells": table_cells,
            "table_rows": rows,
            "width_of_cols": width_of_cols,
            "position": [int(x0), int(y0), int(x1), int(y1)],
            "lines": table_lines,
            "table_cols": cols
        }
        
        return table
    
    def _detect_merged_cells(self, table_data: List[List]) -> Dict[tuple, Dict]:
        """
        检测合并单元格
        
        Args:
            table_data: 二维表格数据
            
        Returns:
            Dict: {(start_row, start_col): {'end_row': int, 'end_col': int, 'text': str}}
        """
        if not table_data:
            return {}
        
        rows = len(table_data)
        cols = len(table_data[0]) if table_data else 0
        merged_cells = {}
        processed = set()
        
        for row_idx in range(rows):
            for col_idx in range(cols):
                if (row_idx, col_idx) in processed:
                    continue
                
                current_cell = table_data[row_idx][col_idx] if col_idx < len(table_data[row_idx]) else None
                current_text = str(current_cell).strip() if current_cell is not None else ""
                
                if not current_text:
                    # 空单元格，跳过
                    processed.add((row_idx, col_idx))
                    continue
                
                # 检测向右合并（水平合并）
                end_col = col_idx
                for c in range(col_idx + 1, cols):
                    if c < len(table_data[row_idx]):
                        next_cell = table_data[row_idx][c]
                        next_text = str(next_cell).strip() if next_cell is not None else ""
                        
                        # 如果相邻单元格为空或者内容相同，可能是合并单元格
                        if not next_text or next_text == current_text:
                            end_col = c
                        else:
                            break
                    else:
                        break
                
                # 检测向下合并（垂直合并）
                end_row = row_idx
                can_merge_down = True
                for r in range(row_idx + 1, rows):
                    # 检查这一行的所有相关列是否都可以合并
                    row_can_merge = True
                    for c in range(col_idx, end_col + 1):
                        if c < len(table_data[r]):
                            cell_below = table_data[r][c]
                            text_below = str(cell_below).strip() if cell_below is not None else ""
                            
                            # 如果下方单元格不为空且内容不同，则不能合并
                            if text_below and text_below != current_text:
                                row_can_merge = False
                                break
                        else:
                            row_can_merge = False
                            break
                    
                    if row_can_merge:
                        end_row = r
                    else:
                        break
                
                # 如果检测到合并单元格（跨越多个行或列）
                if end_row > row_idx or end_col > col_idx:
                    # 收集合并单元格中的所有非空文本
                    merged_texts = []
                    for r in range(row_idx, end_row + 1):
                        for c in range(col_idx, end_col + 1):
                            if r < len(table_data) and c < len(table_data[r]):
                                cell_text = str(table_data[r][c]).strip() if table_data[r][c] is not None else ""
                                if cell_text and cell_text not in merged_texts:
                                    merged_texts.append(cell_text)
                    
                    # 合并文本内容
                    final_text = " ".join(merged_texts) if merged_texts else current_text
                    
                    merged_cells[(row_idx, col_idx)] = {
                        'end_row': end_row,
                        'end_col': end_col,
                        'text': final_text
                    }
                    
                    logger.debug(f"检测到合并单元格: ({row_idx},{col_idx}) -> ({end_row},{end_col}), 内容: '{final_text}'")
                
                # 标记所有相关位置为已处理
                for r in range(row_idx, end_row + 1):
                    for c in range(col_idx, end_col + 1):
                        processed.add((r, c))
        
        return merged_cells
    
    def _detect_merged_cells_pymupdf(self, table_data: List[List]) -> Dict[tuple, Dict]:
        """
        检测合并单元格（PyMuPDF版本，更严格的检测逻辑）
        
        Args:
            table_data: 二维表格数据
            
        Returns:
            Dict: {(start_row, start_col): {'end_row': int, 'end_col': int, 'text': str}}
        """
        if not table_data:
            return {}
        
        rows = len(table_data)
        cols = len(table_data[0]) if table_data else 0
        merged_cells = {}
        processed = set()
        
        for row_idx in range(rows):
            for col_idx in range(cols):
                if (row_idx, col_idx) in processed:
                    continue
                
                current_text = str(table_data[row_idx][col_idx]).strip() if table_data[row_idx][col_idx] else ""
                
                if not current_text:
                    # 空单元格，检查是否可以向右或向下查找非空单元格
                    processed.add((row_idx, col_idx))
                    continue
                
                # 更严格的合并检测：只有完全空的相邻单元格才认为是合并
                end_col = col_idx
                end_row = row_idx
                
                # 检测水平合并
                for c in range(col_idx + 1, cols):
                    next_text = str(table_data[row_idx][c]).strip() if table_data[row_idx][c] else ""
                    if not next_text:  # 空单元格，可能是合并的
                        end_col = c
                    else:
                        break
                
                # 检测垂直合并（仅在有水平合并的基础上）
                if end_col > col_idx:
                    # 检查整个水平合并区域是否都可以向下扩展
                    for r in range(row_idx + 1, rows):
                        can_extend = True
                        for c in range(col_idx, end_col + 1):
                            cell_text = str(table_data[r][c]).strip() if table_data[r][c] else ""
                            if cell_text:  # 如果下方有非空内容，不能合并
                                can_extend = False
                                break
                        
                        if can_extend:
                            end_row = r
                        else:
                            break
                else:
                    # 没有水平合并，只检测垂直合并
                    for r in range(row_idx + 1, rows):
                        next_text = str(table_data[r][col_idx]).strip() if table_data[r][col_idx] else ""
                        if not next_text:  # 空单元格，可能是合并的
                            end_row = r
                        else:
                            break
                
                # 如果检测到合并单元格
                if end_row > row_idx or end_col > col_idx:
                    merged_cells[(row_idx, col_idx)] = {
                        'end_row': end_row,
                        'end_col': end_col,
                        'text': current_text
                    }
                    
                    logger.debug(f"PyMuPDF检测到合并单元格: ({row_idx},{col_idx}) -> ({end_row},{end_col}), 内容: '{current_text}'")
                
                # 标记所有相关位置为已处理
                for r in range(row_idx, end_row + 1):
                    for c in range(col_idx, end_col + 1):
                        processed.add((r, c))
        
        return merged_cells
    
    def _create_char_info_for_text(self, text: str, position: List[int]) -> Dict[str, List]:
        """为文本创建字符级信息（简化版）"""
        char_info = {
            "char_attributes": [],
            "char_candidates": [],
            "char_candidates_score": [],
            "char_centers": [],
            "char_positions": [],
            "char_scores": []
        }
        
        if not text:
            return char_info
        
        x0, y0, x1, y1 = position
        char_width = (x1 - x0) / len(text) if len(text) > 0 else 0
        
        for i, char in enumerate(text):
            if char.strip():
                char_info["char_attributes"].append("normal")
                char_info["char_candidates"].append([char])
                char_info["char_candidates_score"].append([1.0])
                
                char_x0 = x0 + i * char_width
                char_x1 = char_x0 + char_width
                
                center_x = int((char_x0 + char_x1) / 2)
                center_y = int((y0 + y1) / 2)
                char_info["char_centers"].append([center_x, center_y])
                
                char_info["char_positions"].append([
                    int(char_x0), int(y0), int(char_x1), int(y1)
                ])
                
                char_info["char_scores"].append(1.0)
        
        return char_info


class HybridTableExtractor:
    """混合表格提取器 - 结合pdfplumber和PyMuPDF（优化版）"""
    
    def __init__(self, pdf_path: str, page, need_character: bool = False, pdf_plumber_doc=None, engine_mode: str = "hybrid", high_accuracy: bool = False):
        """
        初始化混合提取器

        Args:
            pdf_path: PDF文件路径
            page: PyMuPDF page对象
            need_character: 是否需要字符级信息
            pdf_plumber_doc: 预先打开的pdfplumber文档对象
            engine_mode: 引擎模式 - "hybrid", "pdfplumber", "pymupdf"
            high_accuracy: 是否启用高精度模式（启用Camelot兜底）
        """
        self.pdf_path = pdf_path
        self.page = page  # PyMuPDF page对象
        self.need_character = need_character
        self.pdf_plumber_doc = pdf_plumber_doc
        self.engine_mode = engine_mode
        self.high_accuracy = high_accuracy
        
        # 根据引擎模式决定是否创建pdfplumber提取器
        if engine_mode in ["hybrid", "pdfplumber"] and pdf_plumber_doc:
            self.plumber_extractor = PDFPlumberTableExtractor(
                pdf_path, need_character, pdf_plumber_doc
            )
        else:
            self.plumber_extractor = None

    def find_tables(self) -> List[Dict[str, Any]]:
        """查找表格 - 根据引擎模式选择策略"""
        tables = []

        # 获取页面号
        page_num = self.page.number

        if self.engine_mode == "pdfplumber":
            # 只使用pdfplumber
            if self.plumber_extractor:
                logger.debug("Using pdfplumber-only mode for table detection")
                plumber_tables = self.plumber_extractor.extract_tables_from_page(page_num)
                tables.extend(plumber_tables)

                if self._should_try_camelot(plumber_tables):
                    logger.debug("pdfplumber result sparse, trying Camelot fallback")
                    camelot_tables = self._extract_with_camelot()
                    tables.extend(self._filter_overlapping_tables(camelot_tables, tables))
            else:
                logger.warning("pdfplumber mode requested but extractor not available")

        elif self.engine_mode == "pymupdf":
            # 只使用PyMuPDF
            logger.debug("Using PyMuPDF-only mode for table detection")
            pymupdf_tables = self._extract_with_pymupdf_simple()
            tables.extend(pymupdf_tables)

        else:  # hybrid mode
            # 混合模式：优先使用pdfplumber，PyMuPDF作为补充
            logger.debug("Using hybrid mode for table detection")

            if self.plumber_extractor:
                plumber_tables = self.plumber_extractor.extract_tables_from_page(page_num)

                if plumber_tables:
                    logger.info(f"pdfplumber found {len(plumber_tables)} tables")
                    tables.extend(plumber_tables)

                    # 即使pdfplumber找到了表格，也尝试PyMuPDF的简化方法作为补充
                    logger.debug("Trying PyMuPDF as supplement")
                    pymupdf_tables = self._extract_with_pymupdf_simple()

                    # 过滤掉与pdfplumber表格重叠的PyMuPDF表格
                    filtered_pymupdf_tables = self._filter_overlapping_tables(pymupdf_tables, plumber_tables)
                    if filtered_pymupdf_tables:
                        logger.debug(f"Adding {len(filtered_pymupdf_tables)} additional PyMuPDF tables")
                        tables.extend(filtered_pymupdf_tables)

                    if self._should_try_camelot(plumber_tables):
                        logger.debug("pdfplumber result sparse, trying Camelot fallback")
                        camelot_tables = self._extract_with_camelot()
                        if camelot_tables:
                            filtered = self._filter_overlapping_tables(camelot_tables, tables)
                            if filtered:
                                logger.debug(f"Adding {len(filtered)} Camelot tables")
                                tables.extend(filtered)
                else:
                    # 如果pdfplumber没找到，尝试PyMuPDF方法
                    logger.debug("pdfplumber found no tables, trying PyMuPDF")
                    pymupdf_tables = self._extract_with_pymupdf_simple()
                    tables.extend(pymupdf_tables)

                    if self.high_accuracy:
                        camelot_tables = self._extract_with_camelot()
                        if camelot_tables:
                            filtered = self._filter_overlapping_tables(camelot_tables, tables)
                            if filtered:
                                logger.debug(f"Adding {len(filtered)} Camelot tables")
                                tables.extend(filtered)
            else:
                # 如果没有pdfplumber，回退到PyMuPDF
                logger.debug("pdfplumber not available, using PyMuPDF only")
                pymupdf_tables = self._extract_with_pymupdf_simple()
                tables.extend(pymupdf_tables)

                if self.high_accuracy:
                    camelot_tables = self._extract_with_camelot()
                    if camelot_tables:
                        filtered = self._filter_overlapping_tables(camelot_tables, tables)
                        if filtered:
                            logger.debug(f"Adding {len(filtered)} Camelot tables")
                            tables.extend(filtered)

        return tables
    
    def _extract_with_pymupdf_simple(self) -> List[Dict[str, Any]]:
        """使用简化的PyMuPDF方法提取表格"""
        tables = []

        try:
            # 获取页面中的所有线条
            drawings = self.page.get_drawings()
            
            if not drawings:
                logger.debug("No drawings found on page")
                return tables
            
            # 简单的线条提取
            h_lines = []
            v_lines = []
            
            for item in drawings:
                if item.get('type') == 'l':  # 线条
                    items = item.get('items', [])
                    for points in items:
                        if len(points) >= 2:
                            p1, p2 = points[0], points[1]
                            self._add_line_if_straight(p1, p2, h_lines, v_lines)
                elif item.get('type') == 're':  # 矩形
                    rect_info = item.get('rect')
                    if rect_info:
                        x0, y0, x1, y1 = rect_info
                        # 添加矩形的四条边
                        h_lines.extend([
                            {'y': y0, 'x1': x0, 'x2': x1},
                            {'y': y1, 'x1': x0, 'x2': x1}
                        ])
                        v_lines.extend([
                            {'x': x0, 'y1': y0, 'y2': y1},
                            {'x': x1, 'y1': y0, 'y2': y1}
                        ])
            
            logger.debug(f"Found {len(h_lines)} h-lines, {len(v_lines)} v-lines with PyMuPDF")
            
            if len(h_lines) >= 2 and len(v_lines) >= 2:
                # 简单创建表格
                table = self._create_simple_table_from_lines(h_lines, v_lines)
                if table:
                    tables.append(table)
            
        except Exception as e:
            logger.warning(f"PyMuPDF simple extraction failed: {e}")
        
        return tables

    def _should_try_camelot(self, tables: List[Dict[str, Any]]) -> bool:
        """判断是否需要尝试Camelot兜底"""
        if not CAMELOT_AVAILABLE:
            return False

        # 如果未启用高精度模式，不使用Camelot
        if not self.high_accuracy:
            return False

        if not tables:
            return True

        # 如果所有表格的列数都很少，说明结构可能缺失
        single_col_tables = [t for t in tables if t.get('table_cols', 0) <= 1]
        return len(single_col_tables) == len(tables)

    def _extract_with_camelot(self) -> List[Dict[str, Any]]:
        """使用Camelot提取表格"""
        if not CAMELOT_AVAILABLE:
            return []

        page_num = self.page.number + 1  # Camelot使用1-based页码
        try:
            camelot_tables = camelot.read_pdf(
                self.pdf_path,
                pages=str(page_num),
                flavor="stream",
                strip_text="\n",
                suppress_stdout=True,
            )
        except Exception as e:
            logger.warning(f"Camelot extraction failed on page {page_num}: {e}")
            return []

        results = []
        for idx, cam_table in enumerate(camelot_tables):
            converted = self._convert_camelot_table(cam_table, idx)
            if converted:
                results.append(converted)

        if results:
            logger.debug(f"Camelot extracted {len(results)} tables on page {page_num}")

        return results

    def _convert_camelot_table(self, cam_table, table_idx: int) -> Optional[Dict[str, Any]]:
        """将Camelot表格转换为统一结构"""
        try:
            df = cam_table.df
        except Exception as e:
            logger.warning(f"Camelot table {table_idx} missing dataframe: {e}")
            return None

        if df is None or df.empty:
            return None

        rows, cols = df.shape
        if rows == 0 or cols == 0:
            return None

        # Camelot坐标以左下为原点，需要转换为与PyMuPDF一致的坐标系
        page_height = float(self.page.rect.height)

        def convert_y(value: float) -> float:
            return page_height - value

        # 若Camelot未提供单元格坐标，则根据整体bbox平均切分
        cell_positions = None
        try:
            if cam_table.cells:
                cell_positions = cam_table.cells
        except Exception:
            cell_positions = None

        table_cells = []

        if cell_positions and len(cell_positions) == rows:
            for r in range(rows):
                if len(cell_positions[r]) != cols:
                    cell_positions = None
                    break

        if cell_positions and len(cell_positions) == rows:
            # 根据Camelot提供的坐标构建单元格
            for r in range(rows):
                for c in range(cols):
                    cell_geom = cell_positions[r][c]
                    text = str(df.iat[r, c]).strip()
                    x0 = float(cell_geom.x1)
                    x1 = float(cell_geom.x2)
                    y_top = convert_y(float(cell_geom.y2))
                    y_bottom = convert_y(float(cell_geom.y1))
                    y0, y1 = sorted([y_top, y_bottom])

                    lines = []
                    if text:
                        line = {
                            "angle": 0,
                            "text": text,
                            "direction": 0,
                            "handwritten": 0,
                            "position": [int(x0), int(y0), int(x1), int(y1)],
                            "score": 1.0,
                            "type": "text",
                        }
                        if self.need_character:
                            char_info = self._create_char_info_for_text(text, line["position"])
                            line.update(char_info)
                        lines.append(line)

                    table_cells.append({
                        "start_row": r,
                        "start_col": c,
                        "end_row": r,
                        "end_col": c,
                        "text": text,
                        "borders": {"right": 1, "bottom": 1, "left": 1, "top": 1},
                        "position": [int(x0), int(y0), int(x1), int(y1)],
                        "lines": lines,
                    })
        else:
            # 退化逻辑：使用表格整体bbox平均切分
            try:
                x0, y0, x1, y1 = cam_table._bbox
            except Exception as e:
                logger.warning(f"Camelot table {table_idx} missing bbox: {e}")
                return None

            top = convert_y(float(y1))
            bottom = convert_y(float(y0))
            top, bottom = sorted([top, bottom])

            col_width = (float(x1) - float(x0)) / cols
            row_height = (bottom - top) / rows if rows else 0

            for r in range(rows):
                for c in range(cols):
                    text = str(df.iat[r, c]).strip()
                    cell_x0 = float(x0) + c * col_width
                    cell_x1 = float(cell_x0 + col_width)
                    cell_y0 = top + r * row_height
                    cell_y1 = cell_y0 + row_height

                    lines = []
                    if text:
                        line = {
                            "angle": 0,
                            "text": text,
                            "direction": 0,
                            "handwritten": 0,
                            "position": [int(cell_x0), int(cell_y0), int(cell_x1), int(cell_y1)],
                            "score": 1.0,
                            "type": "text",
                        }
                        if self.need_character:
                            char_info = self._create_char_info_for_text(text, line["position"])
                            line.update(char_info)
                        lines.append(line)

                    table_cells.append({
                        "start_row": r,
                        "start_col": c,
                        "end_row": r,
                        "end_col": c,
                        "text": text,
                        "borders": {"right": 1, "bottom": 1, "left": 1, "top": 1},
                        "position": [int(cell_x0), int(cell_y0), int(cell_x1), int(cell_y1)],
                        "lines": lines,
                    })

        if not table_cells:
            return None

        # 估算行高和列宽
        height_of_rows = []
        width_of_cols = []

        rows_by_index = {}
        cols_by_index = {}
        for cell in table_cells:
            r = cell["start_row"]
            c = cell["start_col"]
            y0, y1 = cell["position"][1], cell["position"][3]
            x0, x1 = cell["position"][0], cell["position"][2]
            rows_by_index.setdefault(r, []).append(abs(y1 - y0))
            cols_by_index.setdefault(c, []).append(abs(x1 - x0))

        for r in range(rows):
            values = rows_by_index.get(r)
            if values:
                height_of_rows.append(int(sum(values) / len(values)))
            else:
                height_of_rows.append(0)

        for c in range(cols):
            values = cols_by_index.get(c)
            if values:
                width_of_cols.append(int(sum(values) / len(values)))
            else:
                width_of_cols.append(0)

        # 计算整体位置
        min_x = min(cell["position"][0] for cell in table_cells)
        min_y = min(cell["position"][1] for cell in table_cells)
        max_x = max(cell["position"][2] for cell in table_cells)
        max_y = max(cell["position"][3] for cell in table_cells)

        return {
            "height_of_rows": height_of_rows,
            "type": "table_with_line",
            "table_cells": table_cells,
            "table_rows": rows,
            "width_of_cols": width_of_cols,
            "position": [min_x, min_y, max_x, max_y],
            "lines": [],
            "table_cols": cols,
        }
    
    def _add_line_if_straight(self, p1, p2, h_lines, v_lines):
        """添加直线到对应列表"""
        if not p1 or not p2:
            return
        
        x1, y1 = p1.get('x', 0), p1.get('y', 0)
        x2, y2 = p2.get('x', 0), p2.get('y', 0)
        
        # 计算长度
        length = ((x2 - x1) ** 2 + (y2 - y1) ** 2) ** 0.5
        if length < 20:
            return
        
        # 判断方向
        if abs(y1 - y2) <= 3:  # 水平线
            h_lines.append({
                'y': (y1 + y2) / 2,
                'x1': min(x1, x2),
                'x2': max(x1, x2)
            })
        elif abs(x1 - x2) <= 3:  # 垂直线
            v_lines.append({
                'x': (x1 + x2) / 2,
                'y1': min(y1, y2),
                'y2': max(y1, y2)
            })
    
    def _create_simple_table_from_lines(self, h_lines, v_lines):
        """从线条创建简单表格"""
        try:
            # 去重并排序
            h_coords = sorted(list(set([line['y'] for line in h_lines])))
            v_coords = sorted(list(set([line['x'] for line in v_lines])))
            
            if len(h_coords) < 2 or len(v_coords) < 2:
                return None
            
            rows = len(h_coords) - 1
            cols = len(v_coords) - 1
            
            # 首先提取所有单元格的文本数据
            cell_data = []
            for row in range(rows):
                row_data = []
                for col in range(cols):
                    x0, y0 = v_coords[col], h_coords[row]
                    x1, y1 = v_coords[col + 1], h_coords[row + 1]
                    
                    # 提取单元格文本
                    cell_text, _ = self._extract_cell_text_pymupdf(x0, y0, x1, y1)
                    row_data.append(cell_text)
                cell_data.append(row_data)
            
            # 检测合并单元格
            merged_cells_data = self._detect_merged_cells_pymupdf(cell_data)
            
            table_cells = []
            table_lines = []  # 注意：表格中的lines应该为空，因为所有文本都在cells中
            processed_cells = set()  # 记录已处理的单元格位置
            
            for row in range(rows):
                for col in range(cols):
                    # 跳过已经被合并单元格处理的位置
                    if (row, col) in processed_cells:
                        continue
                    
                    # 检查是否为合并单元格的起始位置
                    merge_info = merged_cells_data.get((row, col))
                    if merge_info:
                        start_row, start_col = row, col
                        end_row, end_col = merge_info['end_row'], merge_info['end_col']
                        merged_text = merge_info['text']
                        
                        # 计算合并单元格的位置
                        x0, y0 = v_coords[start_col], h_coords[start_row]
                        x1, y1 = v_coords[end_col + 1], h_coords[end_row + 1]
                        
                        # 重新提取合并区域的文本和lines
                        cell_text, cell_lines = self._extract_cell_text_pymupdf(x0, y0, x1, y1)
                        
                        # 标记所有被合并的单元格位置为已处理
                        for r in range(start_row, end_row + 1):
                            for c in range(start_col, end_col + 1):
                                processed_cells.add((r, c))
                    else:
                        # 普通单元格
                        start_row = end_row = row
                        start_col = end_col = col
                        
                        x0, y0 = v_coords[col], h_coords[row]
                        x1, y1 = v_coords[col + 1], h_coords[row + 1]
                        
                        # 提取单元格文本
                        cell_text, cell_lines = self._extract_cell_text_pymupdf(x0, y0, x1, y1)
                        processed_cells.add((row, col))
                    
                    cell = {
                        "start_row": start_row,
                        "start_col": start_col,
                        "end_row": end_row,
                        "end_col": end_col,
                        "text": cell_text,
                        "borders": {"right": 1, "bottom": 1, "left": 1, "top": 1},
                        "position": [int(x0), int(y0), int(x1), int(y1)],
                        "lines": cell_lines
                    }
                    table_cells.append(cell)
            
            return {
                "height_of_rows": [int(h_coords[i+1] - h_coords[i]) for i in range(rows)],
                "type": "table_with_line",
                "table_cells": table_cells,
                "table_rows": rows,
                "width_of_cols": [int(v_coords[i+1] - v_coords[i]) for i in range(cols)],
                "position": [int(min(v_coords)), int(min(h_coords)), 
                           int(max(v_coords)), int(max(h_coords))],
                "lines": table_lines,
                "table_cols": cols
            }
            
        except Exception as e:
            logger.warning(f"Failed to create table from lines: {e}")
            return None
    
    def _extract_cell_text_pymupdf(self, x0, y0, x1, y1):
        """使用PyMuPDF提取单元格文本"""
        import fitz
        
        margin = 2
        rect = fitz.Rect(x0 + margin, y0 + margin, x1 - margin, y1 - margin)
        
        try:
            text_dict = self.page.get_text("dict", clip=rect)
        except:
            return "", []
        
        cell_text = ""
        cell_lines = []
        
        for block in text_dict.get('blocks', []):
            if block.get('type') == 0:
                for line in block.get('lines', []):
                    line_text = ""
                    for span in line.get('spans', []):
                        line_text += span.get('text', '')
                    
                    if line_text.strip():
                        cell_text += line_text.strip() + " "
                        
                        line_bbox = line.get('bbox', [x0, y0, x1, y1])
                        tline = {
                            "angle": 0,
                            "text": line_text.strip(),
                            "direction": 0,
                            "handwritten": 0,
                            "position": [int(x) for x in line_bbox],
                            "score": 1.0,
                            "type": "text"
                        }
                        
                        if self.need_character:
                            char_info = self._create_char_info_for_text(line_text.strip(), tline["position"])
                            tline.update(char_info)
                        
                        cell_lines.append(tline)
        
        return cell_text.strip(), cell_lines
    
    def _filter_overlapping_tables(self, new_tables, existing_tables, overlap_threshold=0.3):
        """过滤掉与已有表格重叠的新表格"""
        if not existing_tables:
            return new_tables
        
        filtered_tables = []
        
        for new_table in new_tables:
            new_pos = new_table['position']
            
            # 检查是否与任何已有表格重叠
            overlaps = False
            for existing_table in existing_tables:
                existing_pos = existing_table['position']
                
                if self._tables_overlap(new_pos, existing_pos, overlap_threshold):
                    overlaps = True
                    break
            
            if not overlaps:
                filtered_tables.append(new_table)
        
        return filtered_tables
    
    def _tables_overlap(self, pos1, pos2, threshold=0.3):
        """检查两个表格是否重叠"""
        # 计算重叠区域
        overlap_x0 = max(pos1[0], pos2[0])
        overlap_y0 = max(pos1[1], pos2[1])
        overlap_x1 = min(pos1[2], pos2[2])
        overlap_y1 = min(pos1[3], pos2[3])
        
        if overlap_x1 <= overlap_x0 or overlap_y1 <= overlap_y0:
            return False  # 没有重叠
        
        # 计算重叠面积
        overlap_area = (overlap_x1 - overlap_x0) * (overlap_y1 - overlap_y0)
        
        # 计算较小表格的面积
        area1 = (pos1[2] - pos1[0]) * (pos1[3] - pos1[1])
        area2 = (pos2[2] - pos2[0]) * (pos2[3] - pos2[1])
        min_area = min(area1, area2)
        
        # 如果重叠面积超过较小表格的阈值比例，认为是重叠
        return overlap_area > min_area * threshold
    
    def _create_char_info_for_text(self, text: str, position: List[int]) -> Dict[str, List]:
        """为文本创建字符级信息（简化版）"""
        char_info = {
            "char_attributes": [],
            "char_candidates": [],
            "char_candidates_score": [],
            "char_centers": [],
            "char_positions": [],
            "char_scores": []
        }
        
        if not text:
            return char_info
        
        x0, y0, x1, y1 = position
        char_width = (x1 - x0) / len(text) if len(text) > 0 else 0
        
        for i, char in enumerate(text):
            if char.strip():
                char_info["char_attributes"].append("normal")
                char_info["char_candidates"].append([char])
                char_info["char_candidates_score"].append([1.0])
                
                char_x0 = x0 + i * char_width
                char_x1 = char_x0 + char_width
                
                center_x = int((char_x0 + char_x1) / 2)
                center_y = int((y0 + y1) / 2)
                char_info["char_centers"].append([center_x, center_y])
                
                char_info["char_positions"].append([
                    int(char_x0), int(y0), int(char_x1), int(y1)
                ])
                
                char_info["char_scores"].append(1.0)
        
        return char_info
