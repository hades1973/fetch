## fetchdata
fetchdata从新浪财经爬取股票日线交易数据。
将要爬取的股票代码与名称写入到/gostock/src/data/stocklist.txt文件内。
stocklist.txt文件格式为：
600030 中信证券
002024 苏宁云商
...

运行
./fetchdata init
将会根据股票名册文件stocklist.txt逐个爬取股票日线，分别存放到/gostock/src/data/stockname文件内。
文件格式：
日期 开盘价 最高价 收盘价 最低价 成交量 成交额 权

运行
./fetchdata update
则更新股票日线数据到当前日期。



