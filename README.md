# pwbook

### 架构说明
pwbook 是 Go 语言写的密码本，可以管理密码同时通过加密保证不被泄露。加密采用 AES 算法 CTR 模式，通过十六位的初始向量和十六位的根密码加密和解密，初始向量由 `const seed` 标记，更改源代码中的 `seed` 值可以获得初始向量不同的软件，之间不能相互访问。输入根密码后会验证结果，验证方法是密码本的 `Right` 字段以 `seed` 为初始向量加密后的值等于 `seed`。密码没有明文储存，因而是安全的。

### 安全性提示
1. 根密码输入时不会明文显示，是安全的；但密码显示后会保持显示在终端，请及时关闭终端以免泄露。
2. 文件读取后不会继续持有文件句柄，因此中途删掉文件不会有感知。退出时保存会覆盖或新建文件。
3. 文件无备份，一旦发生错误，所有修改的内容都将丢失。请及时保存或备份。

### 使用说明
启动请在命令行启动，需要输入文件名参数，否则退出同时显示   
`Invalid command! Usage: pwbook <filename>`   
文件不存在提示将会新建，并提示输入密码，请输入 16 位新的根密码

启动后命令如下
1. `exit` 退出同时询问是否保存
2. `help` 展示帮助页面
3. `list` 列出所有密码项
4. `get <id>` 获取 id 为 <id> 的项
5. `remove <id>` 移除 id 为 <id> 的项
6. `add <describe> <password>` 增加描述为 <describe> 的项
7. `change <id> <password>` 修改 id 为 <id> 的项的密码

### 文件格式
文件格式为 .gob，是 Go 语言专用的二进制格式。文件保存了
1. 用于验证密码的 Right 字段（理论上可有可无，不影响读写）
2. 用于储存数据的 BookItems 字段。这是个列表，每一项都是一个密码描述   
   * id 是每一项唯一的标识，即使项的储存位置变化，id 也永不变化
   * description 是对项的描述，不可更改，要更改需要删除项再新建
   * password 是密码项，可以更改，每次获取和输入修改都需要输入根密码验证

### 发布说明
有默认初始向量的版本，已经进行了交叉编译，有 windows、linux 的 amd64、x86、arm 版本和 macos 的 amd64 版本，请根据需要自行下载。

建议自行更换初始向量后编译，增加可靠性。使用到第三方库 `term`，换代理教程 STFW。
