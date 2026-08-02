#!/usr/bin/env python3
# pdfparser.py - PDF解析器主入口文件（清理版）

import sys
import json
import argparse
import time
import os
from pathlib import Path

# 添加当前目录到路径，以便导入模块
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

# 导入自定义模块
from parsers.pdf_parser import PDFParser
from utils.output_formatter import OutputFormatter
from utils.logger import get_logger

logger = get_logger()


def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='Unified PDF Parser')
    parser.add_argument('pdf_path', help='Path to PDF file')
    parser.add_argument('--need-character', action='store_true',
                       help='Include character-level information')
    parser.add_argument('--extract-images', action='store_true',
                       help='Extract images from PDF')
    parser.add_argument('--page-range', default='all',
                       help='Page range, e.g., "1-5", "1,3,5", "all"')
    parser.add_argument('--output-format', choices=['json', 'text', 'html'], default='json',
                       help='Output format')
    parser.add_argument('--table-engine', choices=['pdfplumber', 'pymupdf', 'hybrid'], default='hybrid',
                       help='Table extraction engine to use')
    parser.add_argument('--high-accuracy', action='store_true',
                       help='Enable high accuracy mode (enables Camelot fallback, slower but more accurate)')

    # 保留这些参数以保持兼容性，但内部不再使用或有默认行为
    parser.add_argument('--merge-lines', action='store_true',
                       help='(Deprecated) Merge adjacent text lines - now handled automatically')
    parser.add_argument('--table-confidence', type=float, default=0.8,
                       help='(Deprecated) Confidence threshold - now handled by engine')
    parser.add_argument('--no-column-detection', action='store_true',
                       help='(Deprecated) Column detection is now automatic')
    parser.add_argument('--min-table-rows', type=int, default=2,
                       help='(Deprecated) Minimum table rows - now handled by engine')
    parser.add_argument('--min-table-cols', type=int, default=2,
                       help='(Deprecated) Minimum table columns - now handled by engine')
    parser.add_argument('--no-pdfplumber', action='store_true',
                       help='(Deprecated) Use --table-engine pymupdf instead')

    args = parser.parse_args()

    # 验证文件路径
    if not Path(args.pdf_path).exists():
        logger.error(f"PDF file '{args.pdf_path}' not found")
        error_result = {
            "message": f"error: PDF file '{args.pdf_path}' not found",
            "code": -1,
            "version": "1.0.0",
            "duration": 0,
            "result": {"pages": []}
        }
        print(json.dumps(error_result, ensure_ascii=False))
        sys.exit(1)

    try:
        logger.info(f"Starting PDF parsing for: {args.pdf_path}")

        # 处理已废弃的参数
        table_engine = args.table_engine
        if args.no_pdfplumber:
            logger.warning("--no-pdfplumber is deprecated, using --table-engine pymupdf")
            table_engine = 'pymupdf'

        # 创建解析器并解析PDF
        pdf_parser = PDFParser(
            pdf_path=args.pdf_path,
            need_character=args.need_character,
            extract_images=args.extract_images,
            page_range=args.page_range,
            table_engine=table_engine,
            high_accuracy=args.high_accuracy
        )

        result = pdf_parser.parse()

        # 使用格式化器输出结果
        formatter = OutputFormatter()
        output = formatter.format_output(result, args.output_format)
        print(output)

        logger.info("PDF parsing completed successfully")

    except Exception as e:
        logger.exception(f"Unexpected error: {str(e)}")
        error_result = {
            "message": f"error: {str(e)}",
            "code": -1,
            "version": "1.0.0",
            "duration": 0,
            "result": {"pages": []}
        }
        # 即使出错也输出JSON格式，便于Go程序解析
        if args.output_format == 'json':
            print(json.dumps(error_result, ensure_ascii=False))
        else:
            formatter = OutputFormatter()
            output = formatter.format_output(error_result, args.output_format)
            print(output)
        sys.exit(1)


if __name__ == "__main__":
    main()