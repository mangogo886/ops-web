# ECharts 下载说明

## 下载 ECharts

请将 ECharts 5.4.3 的 min.js 文件下载到此目录。

### 方法一：使用浏览器下载

1. 访问：https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js
2. 右键点击页面，选择"另存为"
3. 保存为 `echarts.min.js` 到此目录（`static/echarts/`）

### 方法二：使用 PowerShell（Windows）

在项目根目录执行：

```powershell
Invoke-WebRequest -Uri "https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js" -OutFile "static\echarts\echarts.min.js"
```

### 方法三：使用 curl（Linux/Mac）

```bash
curl -o static/echarts/echarts.min.js https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js
```

### 方法四：使用 wget（Linux）

```bash
wget -O static/echarts/echarts.min.js https://cdn.jsdelivr.net/npm/echarts@5.4.3/dist/echarts.min.js
```

## 验证

下载完成后，确保文件路径为：`static/echarts/echarts.min.js`

文件大小约为 700KB 左右。


