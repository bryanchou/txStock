package main

import (
	"./utils"
	"encoding/json"
	"fmt"
	"github.com/axgle/mahonia"
	"github.com/daviddengcn/go-colortext"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	greenBg      = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	whiteBg      = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
	yellowBg     = string([]byte{27, 91, 57, 48, 59, 52, 51, 109})
	redBg        = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	blueBg       = string([]byte{27, 91, 57, 55, 59, 52, 52, 109})
	magentaBg    = string([]byte{27, 91, 57, 55, 59, 52, 53, 109})
	cyanBg       = string([]byte{27, 91, 57, 55, 59, 52, 54, 109})
	green        = string([]byte{27, 91, 51, 50, 109})
	white        = string([]byte{27, 91, 51, 55, 109})
	yellow       = string([]byte{27, 91, 51, 51, 109})
	red          = string([]byte{27, 91, 51, 49, 109})
	blue         = string([]byte{27, 91, 51, 52, 109})
	magenta      = string([]byte{27, 91, 51, 53, 109})
	cyan         = string([]byte{27, 91, 51, 54, 109})
	reset        = string([]byte{27, 91, 48, 109})
	disableColor = false
)

func main() {
	ticker := time.NewTicker(5 * time.Second)
	var wg  sync.WaitGroup
	ch := make(chan os.Signal)

	fileName := GetCurrentDate() +".txt"
	file,err := os.OpenFile(fileName,os.O_RDWR|os.O_CREATE|os.O_APPEND,0644)
	if err!=nil{
		log.Fatalln("读取日志文件失败",err)
	}
	defer file.Close()

	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM,syscall.SIGKILL)
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("child goroutine bootstrap start")
		for {
			select {
			case <- ticker.C:
				fmt.Println("----------------------"+time.Now().Format("2006-01-02 15:04:05"))
				StartStock(file)
			case <- ch:
				fmt.Println("work well .")
				ticker.Stop()
				file.Close()
				return
			}
		}
	}()
	wg.Wait()
}

func StartStock(file *os.File,) {
	defer ct.ResetColor()
	stocks,iniParser := GetAllStock()

	var keys []int
	for k := range stocks {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for k := range keys {
		txs := GetstockInfo(stocks[keys[k]], file)
		GetStockMoney(stocks[keys[k]], &txs)
		txs.CurrentTime = time.Now().Format("2006-01-02 15:04:05")
		txs.saveToFile(file)
		formatStr := iniParser.GetString("output","format")
		if strings.Trim(formatStr," ") == "" {
			fmt.Printf("%-12s C:%.2f H:%.2f L:%.2f O:%.2f riseAndFall:%.2f %.2f%% volumn:%.0f\n",txs.StockName, txs.Price,txs.TheHightest,txs.TheLowest,txs.OpenPrice,txs.RiseAndFall,txs.PricesEtc,txs.Volume)
		} else {
			var consoleStr string
			formatStrs := strings.Split(formatStr," ")
			_,paramValues := GetPrintValue(iniParser,txs)


			if len(paramValues) > 0 {
				for i ,v := range paramValues {
					consoleStr += fmt.Sprintf(formatStrs[i],v)+" "
				}
			}
			if txs.RiseAndFall < 0 {
				ct.ChangeColor(ct.Green,true,ct.Black,false)
				fmt.Println(consoleStr)
				ct.ResetColor()
			} else if txs.RiseAndFall == 0 {
				fmt.Println(consoleStr)
			} else {
				ct.ChangeColor(ct.Red,true,ct.Black,false)
				fmt.Println(consoleStr)
				ct.ResetColor()
			}
		}



	}
}

func ProcessRangeString(str string,l int) string {
	nullStr:=""
	if strings.Trim(str," ") != "" {
		if len(str) < l {
			for i:=0; i<l-len(str);i++ {
				nullStr += nullStr+" "
			}
		}
	}
	return str+nullStr
}

func GetPrintValue(iniParse *utils.IniParser,txsStock TXStock) (string,[]interface{}) {
	secionKeys := iniParse.GetSectionKeys("output")
	secionLen := len(secionKeys)
	colomnKeys := make([]int,secionLen-1)
	if  secionLen > 0 {
		for _ , v := range secionKeys {
			if v=="format" {
				continue
			}
			_k,_ := strconv.ParseUint(v,10,32)
			_v := int(iniParse.GetInt64("output",v))
			if _v > 0 {
				colomnKeys[_k-1] = _v
			}
		}
	}
	//sort.Ints(colomnKeys[:])
	var str string
	columnLen := len(colomnKeys)
	var paramValues = make([]interface{},columnLen)
	for i ,v := range colomnKeys {
		switch v {
		case 1:
			str += txsStock.StockName + ","
			paramValues[i]=txsStock.StockName
			break
		case 2:
			str += txsStock.StockCode+ ","
			paramValues[i]=txsStock.StockCode
			break
		case 3:
			paramValues[i]=txsStock.Price
			s := strconv.FormatFloat(float64(txsStock.Price),'f',2, 32)
			str += s + ","
			break
		case 4:
			paramValues[i]=txsStock.LastPrice
			s := strconv.FormatFloat(float64(txsStock.LastPrice),'f',2, 32)
			str += s + ","
			break
		case 5:
			paramValues[i]=txsStock.OpenPrice
			s := strconv.FormatFloat(float64(txsStock.OpenPrice),'f',2, 32)
			str += s + ","
			break
		case 6:
			paramValues[i]=txsStock.CurrentVolumn
			s := strconv.FormatFloat(float64(txsStock.CurrentVolumn),'f',2, 32)
			str += s + ","
			break
		case 7:
			paramValues[i]=txsStock.CurrentVolumn
			s := strconv.FormatFloat(float64(txsStock.OuterVolumn),'f',2, 32)
			str += s + ","
			break
		case 8:
			paramValues[i]=txsStock.Invol
			s := strconv.FormatFloat(float64(txsStock.Invol),'f',2, 32)
			str += s + ","
			break
		case 9:
			paramValues[i]=txsStock.BuyFirst
			s := strconv.FormatFloat(float64(txsStock.BuyFirst),'f',2, 32)
			str += s + ","
			break
		case 10:
			paramValues[i]=txsStock.BuyFirstVolumn
			s := strconv.FormatFloat(float64(txsStock.BuyFirstVolumn),'f',2, 32)
			str += s + ","
			break
		case 11:
			paramValues[i]=txsStock.BuySecond
			s := strconv.FormatFloat(float64(txsStock.BuySecond),'f',2, 32)
			str += s + ","
			break
		case 12:
			paramValues[i]=txsStock.BuySecondVolumn
			s := strconv.FormatFloat(float64(txsStock.BuySecondVolumn),'f',2, 32)
			str += s + ","
			break
		case 13:
			paramValues[i]=txsStock.BuyThird
			s := strconv.FormatFloat(float64(txsStock.BuyThird),'f',2, 32)
			str += s + ","
			break
		case 14:
			paramValues[i]=txsStock.BuyThirdVolumn
			s := strconv.FormatFloat(float64(txsStock.BuyThirdVolumn),'f',2, 32)
			str += s + ","
			break
		case 15:
			paramValues[i]=txsStock.BuyFourth
			s := strconv.FormatFloat(float64(txsStock.BuyFourth),'f',2, 32)
			str += s + ","
			break
		case 16:
			paramValues[i]=txsStock.BuyFourthVolumn
			s := strconv.FormatFloat(float64(txsStock.BuyFourthVolumn),'f',2, 32)
			str += s + ","
			break
		case 17:
			paramValues[i]=txsStock.BuyFifth
			s := strconv.FormatFloat(float64(txsStock.BuyFifth),'f',2, 32)
			str += s + ","
			break
		case 18:
			paramValues[i]=txsStock.BuyFifthVolumn
			s := strconv.FormatFloat(float64(txsStock.BuyFifthVolumn),'f',2, 32)
			str += s + ","
			break
		case 19:
			paramValues[i]=txsStock.SellFirst
			s := strconv.FormatFloat(float64(txsStock.SellFirst),'f',2, 32)
			str += s + ","
			break
		case 20:
			paramValues[i]=txsStock.SellFirstVolumn
			s := strconv.FormatFloat(float64(txsStock.SellFirstVolumn),'f',2, 32)
			str += s + ","
			break
		case 21:
			paramValues[i]=txsStock.SellSecond
			s := strconv.FormatFloat(float64(txsStock.SellSecond),'f',2, 32)
			str += s + ","
			break
		case 22:
			paramValues[i]=txsStock.SellSecondVolumn
			s := strconv.FormatFloat(float64(txsStock.SellSecondVolumn),'f',2, 32)
			str += s + ","
			break
		case 23:
			paramValues[i]=txsStock.SellThird
			s := strconv.FormatFloat(float64(txsStock.SellThird),'f',2, 32)
			str += s + ","
			break
		case 24:
			paramValues[i]=txsStock.SellThirdVolumn
			s := strconv.FormatFloat(float64(txsStock.SellThirdVolumn),'f',2, 32)
			str += s + ","
			break
		case 25:
			paramValues[i]=txsStock.SellFourth
			s := strconv.FormatFloat(float64(txsStock.SellFourth),'f',2, 32)
			str += s + ","
			break
		case 26:
			paramValues[i]=txsStock.SellFourthVolumn
			s := strconv.FormatFloat(float64(txsStock.SellFourthVolumn),'f',2, 32)
			str += s + ","
			break
		case 27:
			paramValues[i]=txsStock.SellFifth
			s := strconv.FormatFloat(float64(txsStock.SellFifth),'f',2, 32)
			str += s + ","
			break
		case 28:
			paramValues[i]=txsStock.SellFifthVolumn
			s := strconv.FormatFloat(float64(txsStock.SellFifthVolumn),'f',2, 32)
			str += s + ","
			break
		case 29:
			paramValues[i]=txsStock.RecentDealByDeal
			s := strconv.FormatFloat(float64(txsStock.RecentDealByDeal),'f',2, 32)
			str += s + ","
			break
		case 30:
			paramValues[i]=txsStock.TransactionTime
			s:= txsStock.TransactionTime
			str += s + ","
			break
		case 31:
			paramValues[i]=txsStock.RiseAndFall
			s := strconv.FormatFloat(float64(txsStock.RiseAndFall),'f',2, 32)
			str += s + ","
			break
		case 32:
			paramValues[i]=txsStock.PricesEtc
			s := strconv.FormatFloat(float64(txsStock.PricesEtc),'f',2, 32)
			str += s + "%,"
			break
		case 33:
			paramValues[i]=txsStock.TheHightest
			s := strconv.FormatFloat(float64(txsStock.TheHightest),'f',2, 32)
			str += s + ","
			break
		case 34:
			paramValues[i]=txsStock.TheLowest
			s := strconv.FormatFloat(float64(txsStock.TheLowest),'f',2, 32)
			str += s + ","
			break
		case 35:
			paramValues[i]=txsStock.TmpTurnoverRate
			s := strconv.FormatFloat(float64(txsStock.TmpTurnoverRate),'f',2, 32)
			str += s + "%,"
			break
		case 36:
			paramValues[i]=txsStock.Volume
			s := strconv.FormatFloat(float64(txsStock.Volume),'f',2, 32)
			str += s + ","
			break
		case 37:
			paramValues[i]=txsStock.Turnover
			s := strconv.FormatFloat(float64(txsStock.Turnover),'f',2, 32)
			str += s + ","
			break
		case 38:
			paramValues[i]=txsStock.TurnoverRate
			s := strconv.FormatFloat(float64(txsStock.TurnoverRate),'f',2, 32)
			str += s + "%,"
			break
		case 39:
			paramValues[i]=txsStock.TTM
			s := strconv.FormatFloat(float64(txsStock.TTM),'f',2, 32)
			str += s + ","
			break
		case 41:
			paramValues[i]=txsStock.TheHightest2
			s := strconv.FormatFloat(float64(txsStock.TheHightest2),'f',2, 32)
			str += s + ","
			break
		case 42:
			paramValues[i]=txsStock.TheLowest2
			s := strconv.FormatFloat(float64(txsStock.TheLowest2),'f',2, 32)
			str += s + ","
			break
		case 43:
			paramValues[i]=txsStock.Amplitude
			s := strconv.FormatFloat(float64(txsStock.Amplitude),'f',2, 32)
			str += s + ","
			break
		case 44:
			paramValues[i]=txsStock.CirculationMarketValue
			s := strconv.FormatFloat(float64(txsStock.CirculationMarketValue),'f',2, 32)
			str += s + ","
			break
		case 45:
			paramValues[i]=txsStock.TotalMarketValue
			s := strconv.FormatFloat(float64(txsStock.TotalMarketValue),'f',2, 32)
			str += s + ","
			break
		case 46:
			paramValues[i]=txsStock.PriceToBookRatio
			s := strconv.FormatFloat(float64(txsStock.PriceToBookRatio),'f',2, 32)
			str += s + ","
			break
		case 47:
			paramValues[i]=txsStock.PriceLimit
			s := strconv.FormatFloat(float64(txsStock.PriceLimit),'f',2, 32)
			str += s + ","
			break
		case 48:
			paramValues[i]=txsStock.LimitPrice
			s := strconv.FormatFloat(float64(txsStock.LimitPrice),'f',2, 32)
			str += s + ","
			break
		case 49:
			paramValues[i]=txsStock.MainInflow
			s := strconv.FormatFloat(txsStock.MainInflow,'f',2, 32)
			str += s + ","
			break
		case 50:
			paramValues[i]=txsStock.MainOutflow
			s := strconv.FormatFloat(txsStock.MainOutflow,'f',2, 32)
			str += s + ","
			break
		case 51:
			paramValues[i]=txsStock.MainNetInflow
			s := strconv.FormatFloat(txsStock.MainNetInflow,'f',2, 32)
			str += s + ","
			break
		case 52:
			paramValues[i]=txsStock.ProportionOfMainInflow
			s := strconv.FormatFloat(txsStock.ProportionOfMainInflow,'f',2, 32)
			str += s + "%,"
			break
		case 53:
			paramValues[i]=txsStock.TheInflowOfRetailInvestors
			s := strconv.FormatFloat(txsStock.TheInflowOfRetailInvestors,'f',2, 32)
			str += s + ","
			break
		case 54:
			paramValues[i]=txsStock.RetailOutflow
			s := strconv.FormatFloat(txsStock.RetailOutflow,'f',2, 32)
			str += s + ","
			break
		case 55:
			paramValues[i]=txsStock.NetInflowOfRetailInvestors
			s := strconv.FormatFloat(txsStock.NetInflowOfRetailInvestors,'f',2, 32)
			str += s + ","
			break
		case 56:
			paramValues[i]=txsStock.InflowRatioOfRetailInvestors
			s := strconv.FormatFloat(txsStock.InflowRatioOfRetailInvestors,'f',2, 32)
			str += s + ","
			break
		case 57:
			paramValues[i]=txsStock.SumOfCapitalInflowAndOutflow
			s := strconv.FormatFloat(txsStock.SumOfCapitalInflowAndOutflow,'f',2, 32)
			str += s + "%,"
			break
		case 58:
			paramValues[i]=txsStock.CurrentTime
			s := txsStock.CurrentTime
			str += s + ","
			break
		}
	}
	return str[0:len(str)-1],paramValues
}

func GetAllStock() (stocks map[int]string,iniParse *utils.IniParser) {
	ini_parser  := utils.IniParser{}
	iniParse = &ini_parser
	/*dir, _ :=  os.Getwd()
	exPath := filepath.Dir(dir)*/
	exPath ,_ :=filepath.Abs(filepath.Dir("stock.ini"))
	conf_file_name :=  exPath + "/stock.ini"
	if err := ini_parser.Load(conf_file_name); err != nil {
		fmt.Printf("try load config file[%s] error[%s]\n", conf_file_name, err.Error())
		return
	}
	stocks =  make(map[int]string)
	sections := ini_parser.GetAllSection()
	if len(sections) > 0 {
		for index,value := range sections {
			if strings.Contains(value,"output") {
				continue
			}
			keys := ini_parser.GetSectionKeys(sections[index])
			if len(keys) > 0 {
				for _,value1 := range keys {
					stockCode := ini_parser.GetString(value,value1)
					mapKey := value+stockCode
					i, _ := strconv.ParseInt(stockCode, 10, 32)
					stocks[int(i)] = mapKey
				}
			}
		}
	}
	return stocks,iniParse
}

type TXStock struct  {
	//http://qt.gtimg.cn/q=sh600519
	//未知字段 0
	UnknownColumn 	string
	//股票名字 1
	StockName		string
	//股票代码 2
	StockCode		string
	//当前价格 3
	Price			float32
	//昨收 4
	LastPrice		float32
	//今开 5
	OpenPrice		float32
	//成交量（手） 6
	CurrentVolumn	float32
	//外盘 7
	OuterVolumn		float32
	//内盘 8
	Invol			float32
	//买一 9
	BuyFirst		float32
	//买一量（手） 10
	BuyFirstVolumn	float32
	//mai 11
	BuySecond		float32
	//mai 12
	BuySecondVolumn float32
	//mai 13
	BuyThird		float32
	//mai 14
	BuyThirdVolumn	float32
	//mai 15
	BuyFourth		float32
	//mai 16
	BuyFourthVolumn float32
	//mai 17
	BuyFifth		float32
	//mai 18
	BuyFifthVolumn  float32
	//卖 19
	SellFirst		float32
	SellFirstVolumn	float32
	//卖 21
	SellSecond		float32
	SellSecondVolumn	float32
	//卖 23
	SellThird		float32
	SellThirdVolumn	float32
	//卖 25
	SellFourth		float32
	SellFourthVolumn	float32
	//卖 27
	SellFifth		float32
	SellFifthVolumn	float32
	//最近逐笔成交 29
	RecentDealByDeal 	float32
	//时间 30
	TransactionTime		string
	//涨跌 31
	RiseAndFall		float32
	// prices/etc 32 涨跌%
	PricesEtc		float32
	// 最高 33
	TheHightest		float32
	//最低 34
	TheLowest		float32
	//价格/成交量（手）/成交额  35 换手率
	TmpTurnoverRate	float32
    //成交量（手)  36
	Volume			float32
	// 成交额（万） 37
	Turnover		float32
	//换手率    38
	TurnoverRate	float32
	//TTM 市盈率  39
	TTM				float32
	//最高  41
	TheHightest2		float32
	//最低 42
	TheLowest2		float32
	//振幅 amplitude  43
	Amplitude		float32
	//流通市值  44
	CirculationMarketValue float32
	//总市值  45
	TotalMarketValue		float32
	//市净率   46
	PriceToBookRatio		float32
	//涨停价 47
	PriceLimit				float32
	//跌停价 48
	LimitPrice				float32

	// 实时资金流向
	//主力流入  1
	MainInflow			float64
	//主力流出 2
	MainOutflow			float64
	//主力净流入 3
	MainNetInflow		float64
	//主力流入占比 4  主力净流入/资金流入流出总和
	ProportionOfMainInflow		float64
	//散户流入 5
	TheInflowOfRetailInvestors	float64
	//Retail outflow  散户流出 6
	RetailOutflow				float64
	//Net inflow of retail investors 散户净流入 7
	NetInflowOfRetailInvestors	float64
	//散户净流入/资金流入流出总和  散户流入比 8
	InflowRatioOfRetailInvestors float64
	//资金流入流出总和  主力流入+主力流出+散户流入+散户流出 9
	SumOfCapitalInflowAndOutflow	float64
	// 当前时间
	CurrentTime					string
}

func GetStockMoney(stockCode string,txstock *TXStock) {
	// 获取实时资金流向 http://qt.gtimg.cn/q=ff_sh600519
	req, _ := http.NewRequest("GET", "http://qt.gtimg.cn/q=ff_"+stockCode, nil)
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 Edg/85.0.564.4")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println("query topic failed", err.Error())
		log.Fatalln("获取实时资金流向失败。",err)
	}
	defer resp.Body.Close()

	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("query topic failed", err.Error())
		log.Fatalln("获取实时资金流向失败。",err)
	}
	respStr := mahonia.NewDecoder("gbk").ConvertString(string(respByte))
	values  := strings.Split(respStr,"~")
	if len(values) > 1 {
		px,_ :=  strconv.ParseFloat(values[1],64)
		txstock.MainInflow = px
		px,_ =  strconv.ParseFloat(values[2],64)
		txstock.MainOutflow = px
		px,_ =  strconv.ParseFloat(values[3],64)
		txstock.MainNetInflow = px
		px,_ =  strconv.ParseFloat(values[4],64)
		txstock.ProportionOfMainInflow = px
		px,_ =  strconv.ParseFloat(values[5],64)
		txstock.TheInflowOfRetailInvestors = px
		px,_ =  strconv.ParseFloat(values[6],64)
		txstock.RetailOutflow = px
		px,_ =  strconv.ParseFloat(values[7],64)
		txstock.NetInflowOfRetailInvestors = px
		px,_ =  strconv.ParseFloat(values[8],64)
		txstock.InflowRatioOfRetailInvestors = px
		px,_ =  strconv.ParseFloat(values[9],64)
		txstock.SumOfCapitalInflowAndOutflow = px
	}
}

func GetstockInfo(stockCode string,file *os.File) TXStock {
	//1 获取基本信息  http://qt.gtimg.cn/q=sh600519
	txstock := TXStock{}
	req, _ := http.NewRequest("GET", "http://qt.gtimg.cn/q="+stockCode, nil)
	req.Header.Set("User-Agent","Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.83 Safari/537.36 Edg/85.0.564.4")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		fmt.Println("query topic failed", err.Error())
		return txstock
	}
	defer resp.Body.Close()
	respByte, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("query topic failed", err.Error())
		return txstock
	}

	respStr := mahonia.NewDecoder("gbk").ConvertString(string(respByte))
	values  := strings.Split(respStr,"~")
	if len(values) > 1 {
		txstock.StockName = values[1]
		txstock.StockCode = values[2]
		px,_ :=  strconv.ParseFloat(values[3],32)
		txstock.Price = float32(px)
		px,_ =  strconv.ParseFloat(values[4],32)
		txstock.LastPrice = float32(px)
		px,_ =  strconv.ParseFloat(values[5],32)
		txstock.OpenPrice = float32(px)
		px,_ =  strconv.ParseFloat(values[6],32)
		txstock.CurrentVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[7],32)
		txstock.OuterVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[8],32)
		txstock.Invol = float32(px)
		px,_ =  strconv.ParseFloat(values[9],32)
		txstock.BuyFirst = float32(px)
		px,_ =  strconv.ParseFloat(values[10],32)
		txstock.BuyFirstVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[11],32)
		txstock.BuySecond = float32(px)
		px,_ =  strconv.ParseFloat(values[12],32)
		txstock.BuySecondVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[13],32)
		txstock.BuyThird = float32(px)
		px,_ =  strconv.ParseFloat(values[14],32)
		txstock.BuyThirdVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[15],32)
		txstock.BuyFourth = float32(px)
		px,_ =  strconv.ParseFloat(values[16],32)
		txstock.BuyFourthVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[17],32)
		txstock.BuyFifth = float32(px)
		px,_ =  strconv.ParseFloat(values[18],32)
		txstock.BuyFifthVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[19],32)
		txstock.SellFirst = float32(px)
		px,_ =  strconv.ParseFloat(values[20],32)
		txstock.SellFirstVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[21],32)
		txstock.SellSecond = float32(px)
		px,_ =  strconv.ParseFloat(values[22],32)
		txstock.SellSecondVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[23],32)
		txstock.SellThird = float32(px)
		px,_ =  strconv.ParseFloat(values[24],32)
		txstock.SellThirdVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[25],32)
		txstock.SellFourth = float32(px)
		px,_ =  strconv.ParseFloat(values[26],32)
		txstock.SellFourthVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[27],32)
		txstock.SellFifth = float32(px)
		px,_ =  strconv.ParseFloat(values[28],32)
		txstock.SellFifthVolumn = float32(px)
		px,_ =  strconv.ParseFloat(values[29],32)
		txstock.RecentDealByDeal = float32(px)
		//txstock.TransactionTime,_ =  time.Parse("2006-01-02 15:04:05", values[30])
		txstock.TransactionTime = values[30]
		px,_ =  strconv.ParseFloat(values[31],32)
		txstock.RiseAndFall = float32(px)
		px,_ =  strconv.ParseFloat(values[32],32)
		txstock.PricesEtc = float32(px)
		px,_ =  strconv.ParseFloat(values[33],32)
		txstock.TheHightest = float32(px)
		px,_ =  strconv.ParseFloat(values[34],32)
		txstock.TheLowest = float32(px)
		px,_ =  strconv.ParseFloat(values[35],32)
		txstock.TmpTurnoverRate = float32(px)
		px,_ =  strconv.ParseFloat(values[36],32)
		txstock.Volume = float32(px)
		px,_ =  strconv.ParseFloat(values[37],32)
		txstock.Turnover = float32(px)
		px,_ =  strconv.ParseFloat(values[38],32)
		txstock.TurnoverRate = float32(px)
		px,_ =  strconv.ParseFloat(values[39],32)
		txstock.TTM = float32(px)
		//txstock.stockName =  strconv.ParseFloat(values[40],32)
		px,_ =  strconv.ParseFloat(values[41],32)
		txstock.TheHightest2 = float32(px)
		px,_ =  strconv.ParseFloat(values[42],32)
		txstock.TheLowest2 = float32(px)
		px,_ =  strconv.ParseFloat(values[43],32)
		txstock.Amplitude = float32(px)
		px,_ =  strconv.ParseFloat(values[44],32)
		txstock.CirculationMarketValue = float32(px)
		px,_ =  strconv.ParseFloat(values[45],32)
		txstock.TotalMarketValue = float32(px)
		px,_ =  strconv.ParseFloat(values[46],32)
		txstock.PriceToBookRatio = float32(px)
		px,_ =  strconv.ParseFloat(values[47],32)
		txstock.PriceLimit = float32(px)
		px,_ =  strconv.ParseFloat(values[48],32)
		txstock.LimitPrice = float32(px)
	}

	return txstock
}

func GetCurrentDate() string {
	return time.Now().Format("2006-01-02")
}

func (txStock TXStock) saveToFile(file *os.File) {
	//err := os.Mkdir(_dir, os.ModePerm)
//struct转json 首字母大写的才会被转
	jsonBytes, err := json.Marshal(txStock)
	if err != nil {
		fmt.Println(err)
	}
	file.WriteString(string(jsonBytes) +" \r\n")
}
