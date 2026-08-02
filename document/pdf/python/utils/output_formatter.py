# utils/output_formatter.py - 输出格式化器

import json
from typing import Dict, Any


class OutputFormatter:
    """输出格式化器"""
    
    def format_output(self, result: Dict[str, Any], output_format: str) -> str:
        """根据指定格式输出结果"""
        if output_format == 'json':
            return self._format_json(result)
        elif output_format == 'text':
            return self._format_text(result)
        elif output_format == 'html':
            return self._format_html(result)
        else:
            raise ValueError(f"Unsupported output format: {output_format}")
    
    def _format_json(self, result: Dict[str, Any]) -> str:
        """输出JSON格式"""
        return json.dumps(result, ensure_ascii=False, indent=2)
    
    def _format_text(self, result: Dict[str, Any]) -> str:
        """输出纯文本格式"""
        lines = []
        
        if result['code'] != 0:
            lines.append(f"Error: {result['message']}")
            return '\n'.join(lines)
        
        for page_idx, page in enumerate(result['result']['pages']):
            lines.append(f"\n=== Page {page_idx + 1} ===\n")
            
            # 输出表格内容
            for table_idx, table in enumerate(page['tables']):
                if table['type'] == 'plain':
                    # 纯文本内容
                    lines.append(f"\n[Text Block {table_idx + 1}]")
                    for line in table['lines']:
                        lines.append(line['text'])
                else:
                    # 表格内容 - 优先显示cells，避免重复
                    lines.append(f"\n[Table {table_idx + 1}] ({table['type']})")
                    if table['table_cells']:
                        # 有单元格数据时，只显示单元格内容
                        for cell in table['table_cells']:
                            if cell['text'].strip():
                                lines.append(f"  [{cell['start_row']},{cell['start_col']}]: {cell['text']}")
                    elif table['lines']:
                        # 没有单元格数据时，显示行数据（表格检测不完整的情况）
                        lines.append("  [检测到文本但未形成单元格结构]")
                        for line in table['lines']:
                            if line['text'].strip():
                                lines.append(f"  {line['text']}")
        
        return '\n'.join(lines)
    
    def _format_html(self, result: Dict[str, Any]) -> str:
        """输出HTML格式"""
        html = []
        html.append("""<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>PDF Parse Result</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .page { margin-bottom: 30px; border: 1px solid #ddd; padding: 20px; }
        .page-header { font-size: 18px; font-weight: bold; margin-bottom: 10px; }
        table { border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        .text-content { margin: 10px 0; line-height: 1.6; }
        .metadata { color: #666; font-size: 14px; }
        .plain-text { background-color: #f9f9f9; padding: 10px; margin: 10px 0; border-left: 3px solid #007bff; }
    </style>
</head>
<body>
""")
        
        html.append(f'<h1>PDF Parse Result</h1>')
        html.append(f'<div class="metadata">Status: {result["message"]} | Duration: {result["duration"]}ms</div>')
        
        if result['code'] != 0:
            html.append(f'<div class="error">Error: {result["message"]}</div>')
            html.append('</body></html>')
            return '\n'.join(html)
        
        for page_idx, page in enumerate(result['result']['pages']):
            html.append(f'<div class="page">')
            html.append(f'<div class="page-header">Page {page_idx + 1} ({page["width"]}x{page["height"]})</div>')
            
            # 处理表格和文本内容
            for table_idx, table in enumerate(page['tables']):
                if table['type'] == 'plain':
                    # 纯文本内容
                    html.append(f'<div class="plain-text">')
                    html.append(f'<h4>Text Block {table_idx + 1}</h4>')
                    for line in table['lines']:
                        html.append(f'<p>{self._escape_html(line["text"])}</p>')
                    html.append('</div>')
                else:
                    # 表格内容 - 优先显示cells，避免重复
                    html.append(f'<h3>Table {table_idx + 1}</h3>')
                    html.append(f'<p class="metadata">Type: {table["type"]} | Size: {table["table_rows"]}x{table["table_cols"]}</p>')
                    
                    if table['table_cells']:
                        # 有单元格数据时，显示表格
                        html.append('<table>')
                        
                        # 创建表格网格
                        grid = [['' for _ in range(table['table_cols'])] for _ in range(table['table_rows'])]
                        
                        # 填充单元格数据
                        for cell in table['table_cells']:
                            if (cell['start_row'] < table['table_rows'] and 
                                cell['start_col'] < table['table_cols']):
                                grid[cell['start_row']][cell['start_col']] = cell['text']
                        
                        # 生成HTML表格
                        for row in grid:
                            html.append('<tr>')
                            for cell_text in row:
                                html.append(f'<td>{self._escape_html(cell_text)}</td>')
                            html.append('</tr>')
                        
                        html.append('</table>')
                    elif table['lines']:
                        # 没有单元格数据时，显示行数据（表格检测不完整的情况）
                        html.append('<div style="color: #888; font-style: italic;">表格检测到文本但未形成单元格结构：</div>')
                        html.append('<ul>')
                        for line in table['lines']:
                            if line['text'].strip():
                                html.append(f'<li>{self._escape_html(line["text"])}</li>')
                        html.append('</ul>')
            
            html.append('</div>')
        
        html.append('</body></html>')
        return '\n'.join(html)
    
    def _escape_html(self, text: str) -> str:
        """HTML转义"""
        text = str(text)
        text = text.replace('&', '&amp;')
        text = text.replace('<', '&lt;')
        text = text.replace('>', '&gt;')
        text = text.replace('"', '&quot;')
        text = text.replace("'", '&#39;')
        return text