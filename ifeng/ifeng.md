## 获取个股历史交易数据
http://api.finance.ifeng.com/akdaily/?code=sh601989&type=last
数据格式： ['date', 'open', 'high', 'close', 'low', 'volume', 'chg', '%chg', 'ma5', 'ma10', 'ma20', 'vma5', 'vma10', 'vma20', 'turnover']
date 日期
open 开盘价
high 最高价
close 收盘价
low 最低价
volume成交量（手） 
chg 涨跌额 
p_chg 涨跌幅
ma5 5日均价
ma10 10日均价
ma20 20日均价
vma5 5日均量
vma10 10日均量
vma20 20日均量
turnover换手率(指数无此项)

## 五日分时成交记录 
http://api.finance.ifeng.com/aminhis/?code=sz002259&type=five
数据格式:
"0": "sz002259",//股票代码
"1": "10.640",//昨收
"2": "10.560",//开盘
"3": "10.770",//最高
"4": "10.360",//最低
"5": "2017-01-13",//日期
"6": "y",
"record": [
    [
    "2017-01-13 09:30",//分时
    "10.560",//价格
    "-0.75",//涨跌幅
    "726.00",//成交量
    "10.56",//均价
    "1.115"//涨跌
    ],...
]
