# 批量导出CSDN博客
> 批量导出`csnd`博客，并转化为`hexo`博客样式，如果你是用富文本编辑器写的则会导出`html`样式

注：有些文章可能获取不到造成进度条无法达到100%，如果走到90%多，走不动了，直接取消即可

# Quick start
```bash
go run main.go -username 你的csdn用户名 -cookie 你csdn的cookie -page 1
```
> page不写，默认为下载全部页

# Demo

```bash
go run main.go -username "junmoxi" -cookie "UserName=junmoxi; UserToken=c3c29cca48be43c4884fe36d052d5851"
```
> 如果想下载别人的文章，那么将`username`更换为别人的即可，cookie还是用你的

# cookie获取
![](cookie.png)

# 关注
> 如果对你有所帮助，请给个star，你的支持是我最大的动力，欢迎关注我微信公众号，一起学习Go语言

![](weixin.png)