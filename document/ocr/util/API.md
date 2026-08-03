# OCR Utility API（Experimental）

`document/ocr/util`提供OCR cache和pipeline使用的两个identity helper：

- `HashFileContents(path)`读取指定文件并返回内容hash；读取失败返回错误；
- `HashPath(path, suffix)`对字符串identity生成稳定路径名，不读取文件。

本包不做路径授权、symlink校验或workspace发现。Host必须先验证调用方提供的路径；hash只用于
cache/identity，不是数字签名或安全完整性证明。
