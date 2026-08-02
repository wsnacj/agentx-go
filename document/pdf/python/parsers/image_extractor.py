# parsers/image_extractor.py - 更新的图片提取器

import base64
import fitz  # PyMuPDF
import sys
import os
import subprocess
import tempfile
import shutil
from typing import List, Dict, Any
from pathlib import Path
import glob

# 添加父目录到路径，以便导入utils模块
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from utils.logger import get_logger

logger = get_logger()


class ImageExtractor:
    """图片提取器 - 更新版本，集成多种提取方法"""

    def __init__(self, doc, pdf_path=None):
        self.doc = doc
        self.pdf_path = pdf_path

        # 如果有PDF路径且FixedImageExtractor可用，使用增强版本
        if pdf_path and FixedImageExtractor:
            self.fixed_extractor = FixedImageExtractor(doc, pdf_path)
            self.use_fixed_extractor = True
            logger.debug("使用增强版图片提取器")
        else:
            self.fixed_extractor = None
            self.use_fixed_extractor = False
            logger.debug("使用基础版图片提取器")

    def extract_images(self, page_num):
        """提取页面中的图片"""
        if self.use_fixed_extractor:
            # 使用增强版提取器
            return self.fixed_extractor.extract_images(page_num)
        else:
            # 使用原有的基础提取器
            return self._extract_images_basic(page_num)

    def _extract_images_basic(self, page_num):
        """基础版图片提取（原有逻辑，作为备用）"""
        page = self.doc[page_num]
        images = []

        try:
            image_list = page.get_images()
            logger.debug(f"基础提取器找到 {len(image_list)} 个图片引用")
        except Exception as e:
            logger.warning(f"基础提取器失败: {e}")
            return []

        for img_index, img in enumerate(image_list):
            try:
                # 获取图片数据
                xref = img[0]

                try:
                    pix = fitz.Pixmap(self.doc, xref)

                    # 应用尺寸过滤策略
                    if hasattr(self, 'fixed_extractor') and self.fixed_extractor:
                        if not self.fixed_extractor._should_extract_image_by_size(
                            pix.width, pix.height, 0):  # 暂时用0作为size_bytes
                            logger.debug(f"跳过小尺寸图片: {img_index} ({pix.width}x{pix.height})")
                            pix = None
                            continue

                    if pix.n - pix.alpha < 4:  # GRAY or RGB
                        img_data = pix.tobytes("png")
                    else:  # CMYK
                        pix1 = fitz.Pixmap(fitz.csRGB, pix)
                        img_data = pix1.tobytes("png")
                        pix1 = None

                    # 应用数据大小过滤策略
                    if hasattr(self, 'fixed_extractor') and self.fixed_extractor:
                        if not self.fixed_extractor._should_extract_image_by_data(img_data, 'png'):
                            logger.debug(f"跳过小图片文件: {img_index} (size: {len(img_data)} bytes)")
                            pix = None
                            continue

                    # 转换为base64
                    img_base64 = base64.b64encode(img_data).decode()

                    # 获取图片位置（可能失败）
                    try:
                        img_rect = page.get_image_bbox(img)
                        position = [int(img_rect.x0), int(img_rect.y0),
                                   int(img_rect.x1), int(img_rect.y1)]
                    except:
                        position = [0, 0, pix.width, pix.height]

                    images.append({
                        "index": img_index,
                        "position": position,
                        "width": int(pix.width),
                        "height": int(pix.height),
                        "data": img_base64,
                        "format": "png",
                        "size_bytes": len(img_data),
                        "extraction_method": "basic_pymupdf"
                    })

                    pix = None

                except Exception as e:
                    logger.warning(f"基础提取器处理图片 {img_index} 失败: {e}")
                    continue

            except Exception as e:
                logger.warning(f"基础提取器提取图片 {img_index} 失败: {e}")
                continue

        return images


class FixedImageExtractor:
    """修复后的图片提取器，优先使用pdfimages"""

    def __init__(self, doc, pdf_path):
        self.doc = doc
        self.pdf_path = pdf_path

    def extract_images(self, page_num):
        """提取页面中的图片，使用多种方法"""
        logger.debug(f"开始提取第 {page_num + 1} 页的图片")

        images = []

        # 方法1: 优先使用pdfimages（最可靠）
        pdfimages_results = self._extract_with_pdfimages(page_num)
        if pdfimages_results:
            images.extend(pdfimages_results)
            logger.debug(f"pdfimages 提取了 {len(pdfimages_results)} 张图片")

        # 方法2: 如果pdfimages失败或没有找到图片，尝试PyMuPDF
        if not images:
            pymupdf_results = self._extract_with_pymupdf(page_num)
            if pymupdf_results:
                images.extend(pymupdf_results)
                logger.debug(f"PyMuPDF 提取了 {len(pymupdf_results)} 张图片")

        # # 方法3: 如果仍然没有图片，渲染页面
        # if not images:
        #     rendered_result = self._extract_by_rendering(page_num)
        #     if rendered_result:
        #         images.append(rendered_result)
        #         logger.debug(f"页面渲染生成 1 张图片")

        logger.info(f"第 {page_num + 1} 页共提取 {len(images)} 张图片")
        return images

    def _extract_with_pdfimages(self, page_num):
        """使用pdfimages提取图片"""
        try:
            # 检查pdfimages是否可用
            if not self._command_exists('pdfimages'):
                logger.debug("pdfimages 命令不可用")
                return []

            # 创建临时目录
            with tempfile.TemporaryDirectory() as temp_dir:
                output_prefix = os.path.join(temp_dir, f"page_{page_num + 1}_")

                # 构建pdfimages命令
                cmd = [
                    'pdfimages',
                    '-f', str(page_num + 1),  # 起始页（1-based）
                    '-l', str(page_num + 1),  # 结束页（1-based）
                    '-all',  # 提取所有图片格式
                    self.pdf_path,
                    output_prefix
                ]

                # 执行命令
                result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)

                if result.returncode != 0:
                    logger.debug(f"pdfimages 执行失败: {result.stderr}")
                    return []

                # 查找生成的图片文件
                image_files = glob.glob(f"{output_prefix}*")
                images = []

                for i, image_file in enumerate(sorted(image_files)):
                    try:
                        image_info = self._process_pdfimages_file(image_file, i, page_num)
                        if image_info:
                            images.append(image_info)
                    except Exception as e:
                        logger.warning(f"处理pdfimages输出文件失败 {image_file}: {e}")

                return images

        except subprocess.TimeoutExpired:
            logger.warning("pdfimages 执行超时")
            return []
        except Exception as e:
            logger.warning(f"pdfimages 提取失败: {e}")
            return []

    def _process_pdfimages_file(self, image_file, index, page_num):
        """处理pdfimages生成的图片文件"""
        try:
            # 获取文件扩展名
            file_ext = os.path.splitext(image_file)[1].lower().lstrip('.')

            # 跳过参数文件
            if file_ext == 'params':
                logger.debug(f"跳过参数文件: {image_file}")
                return None

            # 处理CCITT格式文件
            if file_ext == 'ccitt':
                return self._process_ccitt_file(image_file, index, page_num)

            # 读取图片文件
            with open(image_file, 'rb') as f:
                image_data = f.read()

            # 应用图片过滤策略
            if not self._should_extract_image_by_data(image_data, file_ext):
                logger.debug(f"跳过小图片文件: {image_file} (size: {len(image_data)} bytes)")
                return None

            if not file_ext:
                # 尝试根据文件内容判断格式
                if image_data.startswith(b'\xff\xd8\xff'):
                    file_ext = 'jpg'
                elif image_data.startswith(b'\x89PNG'):
                    file_ext = 'png'
                elif image_data.startswith(b'GIF'):
                    file_ext = 'gif'
                else:
                    file_ext = 'unknown'

            # 获取图片尺寸（尝试多种方法）
            width, height = self._get_image_dimensions(image_data, file_ext)

            # 应用尺寸过滤策略
            if not self._should_extract_image_by_size(width, height, len(image_data)):
                logger.debug(f"跳过小尺寸图片: {image_file} ({width}x{height}, {len(image_data)} bytes)")
                return None

            # 转换为base64
            img_base64 = base64.b64encode(image_data).decode()

            # 估算位置（pdfimages不提供位置信息，使用默认值）
            position = [0, 0, width, height]

            return {
                "index": index,
                "position": position,
                "width": width,
                "height": height,
                "data": img_base64,
                "format": file_ext,
                "size_bytes": len(image_data),
                "extraction_method": "pdfimages"
            }

        except Exception as e:
            logger.warning(f"处理图片文件失败 {image_file}: {e}")
            return None

    def _get_image_dimensions(self, image_data, file_ext):
        """获取图片尺寸"""
        try:
            # 方法1: 使用PIL
            try:
                from PIL import Image
                import io
                with Image.open(io.BytesIO(image_data)) as img:
                    return img.size
            except ImportError:
                pass
            except Exception:
                pass

            # 方法2: 简单的文件头解析
            if file_ext == 'png' and len(image_data) > 24:
                # PNG格式
                if image_data[12:16] == b'IHDR':
                    width = int.from_bytes(image_data[16:20], 'big')
                    height = int.from_bytes(image_data[20:24], 'big')
                    return width, height
            elif file_ext in ['jpg', 'jpeg'] and len(image_data) > 10:
                # JPEG格式的简单解析（不完整但够用）
                # 这里只是示例，实际JPEG解析比较复杂
                pass

            # 默认尺寸
            return 100, 100

        except Exception:
            return 100, 100

    def _extract_with_pymupdf(self, page_num):
        """使用PyMuPDF提取图片（备用方法）"""
        try:
            page = self.doc[page_num]
            images = []

            # 尝试不同的参数组合
            try:
                # 方法1: 不使用full参数
                image_list = page.get_images()
                logger.debug(f"get_images() 找到 {len(image_list)} 个图片引用")
            except Exception as e:
                logger.debug(f"get_images() 失败: {e}")
                return []

            for img_index, img in enumerate(image_list):
                try:
                    image_info = self._extract_pymupdf_image(page, img, img_index, page_num)
                    if image_info:
                        images.append(image_info)
                except Exception as e:
                    logger.warning(f"PyMuPDF提取图片 {img_index} 失败: {e}")

            return images

        except Exception as e:
            logger.warning(f"PyMuPDF提取失败: {e}")
            return []

    def _extract_pymupdf_image(self, page, img, img_index, page_num):
        """使用PyMuPDF提取单个图片"""
        try:
            # 获取图片引用
            xref = img[0]

            # 尝试提取图片
            try:
                # 方法1: 使用extract_image
                if hasattr(self.doc, 'extract_image'):
                    base_image = self.doc.extract_image(xref)
                    image_bytes = base_image["image"]
                    image_ext = base_image["ext"]
                    width = base_image.get("width", 0)
                    height = base_image.get("height", 0)
                else:
                    raise AttributeError("extract_image not available")
            except:
                # 方法2: 使用Pixmap
                pix = fitz.Pixmap(self.doc, xref)
                if pix.n - pix.alpha < 4:  # GRAY or RGB
                    image_bytes = pix.tobytes("png")
                    image_ext = "png"
                else:  # CMYK
                    pix1 = fitz.Pixmap(fitz.csRGB, pix)
                    image_bytes = pix1.tobytes("png")
                    image_ext = "png"
                    pix1 = None

                width = pix.width
                height = pix.height
                pix = None

            # 检查图片大小
            if len(image_bytes) < 100:  # 太小的图片跳过
                return None

            # 获取位置信息
            try:
                # 注意：这里可能会失败
                image_bbox = page.get_image_bbox(img)
                position = [int(image_bbox.x0), int(image_bbox.y0),
                           int(image_bbox.x1), int(image_bbox.y1)]
            except:
                # 使用默认位置
                position = [0, 0, width if width > 0 else 100, height if height > 0 else 100]

            # 转换为base64
            img_base64 = base64.b64encode(image_bytes).decode()

            return {
                "index": img_index,
                "position": position,
                "width": width if width > 0 else 100,
                "height": height if height > 0 else 100,
                "data": img_base64,
                "format": image_ext,
                "size_bytes": len(image_bytes),
                "extraction_method": "pymupdf"
            }

        except Exception as e:
            logger.debug(f"PyMuPDF提取图片失败: {e}")
            return None

    def _extract_by_rendering(self, page_num):
        """渲染页面为图片（最后的备用方案）"""
        try:
            page = self.doc[page_num]

            # 渲染页面
            mat = fitz.Matrix(1.5, 1.5)  # 1.5倍缩放
            pix = page.get_pixmap(matrix=mat, alpha=False)

            # 转换为PNG
            img_data = pix.tobytes("png")

            # 检查大小
            if len(img_data) < 1000:
                return None

            # 转换为base64
            img_base64 = base64.b64encode(img_data).decode()

            return {
                "index": 999,  # 特殊索引表示渲染图片
                "position": [0, 0, int(page.rect.width), int(page.rect.height)],
                "width": pix.width,
                "height": pix.height,
                "data": img_base64,
                "format": "png",
                "size_bytes": len(img_data),
                "extraction_method": "page_render"
            }

        except Exception as e:
            logger.warning(f"页面渲染失败: {e}")
            return None

    def _command_exists(self, command):
        """检查命令是否存在"""
        return shutil.which(command) is not None

    def _should_extract_image_by_size(self, width, height, size_bytes):
        """根据尺寸和大小判断是否应该提取图片"""
        MIN_AREA = 50 * 50  # 最小2500像素面积
        MIN_WIDTH = 30      # 最小宽度30像素
        MIN_HEIGHT = 30     # 最小高度30像素
        MIN_BYTES = 500     # 最小500字节

        area = width * height

        # 面积太小的跳过
        if area < MIN_AREA:
            return False

        # 单边尺寸太小的跳过（可能是线条或装饰）
        if width < MIN_WIDTH or height < MIN_HEIGHT:
            return False

        # 文件太小的跳过
        if size_bytes < MIN_BYTES:
            return False

        return True

    def _should_extract_image_by_data(self, image_data, file_ext):
        """根据图片数据判断是否应该提取"""
        MIN_BYTES = 500  # 最小500字节

        # 文件太小的跳过
        if len(image_data) < MIN_BYTES:
            return False

        # 某些格式的特殊处理
        if file_ext in ['png'] and len(image_data) < 1000:
            # PNG文件如果小于1KB通常是图标或装饰
            return False

        return True

    def _process_ccitt_file(self, ccitt_file, index, page_num):
        """处理CCITT格式文件"""
        try:
            # 查找对应的params文件
            base_name = os.path.splitext(ccitt_file)[0]
            params_file = base_name + '.params'

            if not os.path.exists(params_file):
                logger.warning(f"找不到对应的params文件: {params_file}")
                return None

            # 读取CCITT数据
            with open(ccitt_file, 'rb') as f:
                ccitt_data = f.read()

            # 应用大小过滤
            if not self._should_extract_image_by_data(ccitt_data, 'ccitt'):
                logger.debug(f"跳过小CCITT文件: {ccitt_file} (size: {len(ccitt_data)} bytes)")
                return None

            # 尝试转换CCITT为可用格式
            converted_data = self._convert_ccitt_to_png(ccitt_file, params_file)
            if not converted_data:
                logger.warning(f"无法转换CCITT文件: {ccitt_file}")
                return None

            # 获取转换后的图片尺寸
            width, height = self._get_image_dimensions(converted_data, 'png')

            # 应用尺寸过滤
            if not self._should_extract_image_by_size(width, height, len(converted_data)):
                logger.debug(f"跳过小尺寸CCITT图片: {ccitt_file} ({width}x{height})")
                return None

            # 转换为base64
            img_base64 = base64.b64encode(converted_data).decode()

            return {
                "index": index,
                "position": [0, 0, width, height],
                "width": width,
                "height": height,
                "data": img_base64,
                "format": "png",
                "size_bytes": len(converted_data),
                "extraction_method": "pdfimages_ccitt"
            }

        except Exception as e:
            logger.warning(f"处理CCITT文件失败 {ccitt_file}: {e}")
            return None

    def _convert_ccitt_to_png(self, ccitt_file, params_file):
        """将CCITT文件转换为PNG格式"""
        try:
            # 读取参数文件
            with open(params_file, 'r') as f:
                params = f.read().strip()

            # 这里可以实现CCITT到PNG的转换
            # 由于复杂性，暂时跳过CCITT文件
            logger.debug(f"暂时跳过CCITT文件转换: {ccitt_file}")
            return None

        except Exception as e:
            logger.debug(f"CCITT转换失败: {e}")
            return None
