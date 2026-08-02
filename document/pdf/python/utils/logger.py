# utils/logger.py - 日志工具

import logging
import os
import sys
from pathlib import Path


class PDFParserLogger:
    """PDF解析器专用日志器"""
    
    _instance = None
    _logger = None
    
    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(PDFParserLogger, cls).__new__(cls)
        return cls._instance
    
    def __init__(self):
        if self._logger is None:
            self._setup_logger()
    
    def _setup_logger(self):
        """设置日志器"""
        # 创建日志目录
        log_dir = Path(os.getenv("PDFPARSER_LOG_DIR", "./data/tmp/pdfparser/log"))
        log_dir.mkdir(parents=True, exist_ok=True)
        
        # 日志文件路径
        log_file = log_dir / "pdfparser.log"
        
        # 创建日志器
        self._logger = logging.getLogger('pdfparser')
        self._logger.setLevel(logging.DEBUG)
        
        # 避免重复添加handler
        if not self._logger.handlers:
            # 文件处理器
            file_handler = logging.FileHandler(log_file, encoding='utf-8')
            file_handler.setLevel(logging.DEBUG)
            
            # 格式化器
            formatter = logging.Formatter(
                '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
            )
            file_handler.setFormatter(formatter)
            
            # 添加处理器
            self._logger.addHandler(file_handler)
            
            # 如果设置了调试环境变量，也输出到stderr（不是stdout）
            if os.getenv('PDF_PARSER_DEBUG', '').lower() in ('1', 'true', 'yes'):
                stderr_handler = logging.StreamHandler(sys.stderr)
                stderr_handler.setLevel(logging.DEBUG)
                stderr_handler.setFormatter(formatter)
                self._logger.addHandler(stderr_handler)
    
    def debug(self, message):
        """调试日志"""
        self._logger.debug(message)
    
    def info(self, message):
        """信息日志"""
        self._logger.info(message)
    
    def warning(self, message):
        """警告日志"""
        self._logger.warning(message)
    
    def error(self, message):
        """错误日志"""
        self._logger.error(message)
    
    def exception(self, message):
        """异常日志（包含堆栈信息）"""
        self._logger.exception(message)


# 全局日志器实例
logger = PDFParserLogger()


def get_logger():
    """获取日志器实例"""
    return logger
