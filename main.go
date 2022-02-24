package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	furaApi = "https://neofura.ngd.network/"
	account = "7aMFk2CEo33Sb9R1a1W6q2HxZ1Z97hbLedDy3ZgHJ4i6"
	coinMarketPlace = "https://pro-api.coinmarketcap.com/v2/tools/price-conversion"
)
type Tx struct {
	TxHash string
	TimeStamp  string
	Height string
	Address string
	Token0 string
	Token0Price string
	Token1 string
	Token1Price string
	Token2 string
	Token2Price string
	Token3 string
	Token3Price string
	SysFee string
	NetFee string
}

type Price struct {
	TimeStamp int `json:"timeStamp"`
	Price float64 `json:"price"`
}

type Txs []Tx

var (
	gasPrice map[int]float64
	neoPrice map[int]float64
	ontPrice map[int]float64
	flmPrice map[int]float64
)



func main (){
	var address = flag.String("addr","NhED7joLUgTsdPAp3AjgXmPWDu6ZFA45ww","address")
	var limit = flag.String("l","1000","limit ")
	var skip = flag.String("s","910","skip")


	gasPrice =readCSV("GASBTC-trades-2022-01.csv")
	fmt.Println(gasPrice)
	neoPrice = readCSV("NEOUSDT-trades-2022-01.csv")
	fmt.Println(neoPrice)
	ontPrice = readCSV("ONTUSDT-trades-2022-01.csv")
	fmt.Println(ontPrice)
	flmPrice = readCSV("FLMUSDT-trades-2022-01.csv")
	fmt.Println(flmPrice)
	fmt.Println("Let's go")
	analyseAddressTxs(*address,*limit,*skip)
	fmt.Println(setPrice("GAS", 1643503254))
	fmt.Println(getPriceAPI("BTC",strconv.Itoa(1643573553)))

}
func analyseAddressTxs(address string, limit string , skip string){
	client := &http.Client{}
	//var jsonStrOri string = "{\"jsonrpc\":\"2.0\",\"method\":\"GetRawTransactionByAddress\",\"params\":{\"Address\":\""+address+"\",\"Limit\":500,\"Skip\":1000},\"id\":1\n}"
	var jsonStrOri string = "{\"jsonrpc\":\"2.0\",\"method\":\"GetRawTransactionByAddress\",\"params\":{\"Address\":\""+address+"\",\"Limit\":"+limit+",\"Skip\":"+skip+"},\"id\":1\n}"
	//var jsonStrOri string = "{\"jsonrpc\":\"2.0\",\"method\":\"GetRawTransactionByAddress\",\"params\":{\"Address\":\""+address+"\"},\"id\":1\n}"
	var jsonStr = []byte(jsonStrOri)

	req, err := http.NewRequest("GET",furaApi, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}

	req.Header.Set("Accepts", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	_ = gjson.Get(string(respBody),"result.totalCount").Int()
	queryCount := gjson.Get(string(respBody),"result.result.#").Int()
	fmt.Println(queryCount)

	var txs = Txs{}

	for i := 0; i < int(queryCount); i++ {
		fmt.Println("go to transfer")
		var tx = Tx{}
		var timePath string = "result.result." +strconv.Itoa(i)+".blocktime"
		timeStamp := gjson.Get(string(respBody),timePath).Int()
		if timeStamp < 1640995213011 || timeStamp >1643673447039 {
			continue
		}
		tx.TimeStamp = strconv.FormatInt(timeStamp,10)
		fmt.Println("TimeStamp:",timeStamp)
		var hashPath  string= "result.result."+strconv.Itoa(i)+".hash"
		txHash := gjson.Get(string(respBody),hashPath).String()
		tx.TxHash = txHash
		fmt.Println("TxHash:",txHash)
		var heightPath string = "result.result."+strconv.Itoa(i)+".blockIndex"
		height := gjson.Get(string(respBody),heightPath).String()
		tx.Height = height
		fmt.Println("BlockHeight:",height)

		var netfeePath string = "result.result." +strconv.Itoa(i)+".netfee"
		netfee := math.Pow(10,-8)*gjson.Get(string(respBody),netfeePath).Float()
		tx.NetFee = fmt.Sprint(netfee)
		fmt.Println("NetFee:",netfee)
		var sysfeePath string = "result.result." +strconv.Itoa(i)+".sysfee"
		sysfee := math.Pow(10,-8)*gjson.Get(string(respBody),sysfeePath).Float()
		tx.SysFee = fmt.Sprint(sysfee)
		fmt.Println("Sysfee:",sysfee)
		var detailStrOri string = "{\"jsonrpc\":\"2.0\",\"method\":\"GetNep17TransferByTransactionHash\",\"params\":{\"TransactionHash\":\""+txHash+"\"},\"id\":1\n}"
		var detailStr = []byte(detailStrOri)
		req, err := http.NewRequest("GET",furaApi, bytes.NewBuffer(detailStr))
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}

		req.Header.Set("Accepts", "application/json")
		respDetail, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		respBodyDetail, _ := ioutil.ReadAll(respDetail.Body)
		var transferCount = gjson.Get(string(respBodyDetail), "result.result.#").Int()
		for i := 0; i < int(transferCount); i++ {
			var valuePath  string= "result.result."+strconv.Itoa(i)+".value"
			value := gjson.Get(string(respBodyDetail),valuePath).Float()
			var addressPath string = "result.result."+strconv.Itoa(i)+".from"
			addr := gjson.Get(string(respBodyDetail),addressPath).String()
			fmt.Println("Transfer ",i,": Address:",address)
			var decimalsPath string = "result.result."+strconv.Itoa(i)+".decimals"
			decimals := gjson.Get(string(respBodyDetail),decimalsPath).Int()
			fmt.Println("Transfer ",i,": Decimal:",decimals)
			var symbolPath string = "result.result."+strconv.Itoa(i)+".symbol"
			symbol := gjson.Get(string(respBodyDetail),symbolPath).String()
			fmt.Println("Transfer ",i,": Symbol:",symbol)
			amount := math.Pow(10, float64((-1) *decimals)) *value
			fmt.Println("Transfer ",i,": Amount:",amount)
			intTime, err := strconv.Atoi(strconv.FormatInt(timeStamp,10)[0:10])
			switch i {
			case 0:
				tx.Token0 = fmt.Sprint(amount) +symbol
				tx.Address = addr

				if err != nil {
					fmt.Println(err)
				}
				tx.Token0Price = setPrice(symbol,intTime)

			case 1:
				tx.Token1 = fmt.Sprint(amount) +symbol
				tx.Token1Price = setPrice(symbol,intTime)

			case 2:
				tx.Token2 = fmt.Sprint(amount) +symbol
				tx.Token2Price = setPrice(symbol,intTime)
			case 3:
				tx.Token3 = fmt.Sprint(amount) +symbol
				tx.Token3Price = setPrice(symbol,intTime)
			default:
			}
		}
		txs = append(txs,tx)
		fmt.Println(tx)

	}
	fmt.Println(string(respBody))
	filename := address+".csv"
	csvFile ,err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()
	err = txs.ToCSV(csvFile)
	if err != nil {
		panic(err)
	}
}
func getPriceAPI(token string,timeStamp string) float64{
	var api string = "https://ftx.com/api/markets/"+token +"/USDT/candles"
	client := &http.Client{}
	req, err := http.NewRequest("GET",api, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	q := url.Values{}
	q.Add("resolution", strconv.Itoa(15))
	q.Add("start_time",timeStamp)
	intTime, err := strconv.Atoi(timeStamp)
	if err != nil {
		fmt.Println(err)
	}
	q.Add("end_time",strconv.Itoa(intTime+15))
	req.Header.Set("Accepts","application/json")
	req.Header.Add("X-CMC_PRO_API_KEY", "2be37802-e3cc-4a4d-8418-f1a39ce0f613")
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	result:= gjson.Get(string(respBody),"result.0.open").Float()
	fmt.Println(string(respBody))
	return  result

}
func setPrice(symbol string, timeStamp int) string{
	if symbol == "GAS" {
		if v,ok := gasPrice[timeStamp];!ok{
			for j := 0; j < 999; j++ {
				if n,yes := gasPrice[timeStamp+j+1] ; yes {
					btc := getPriceAPI("BTC",strconv.Itoa(timeStamp))
					return fmt.Sprint(n *btc)

				}
			}
			for j := 0; j < -999; j-- {
				if n,yes := gasPrice[timeStamp-1+j] ; yes {
					btc := getPriceAPI("BTC",strconv.Itoa(timeStamp))
					return fmt.Sprint(n *btc)

				}
			}
		} else {
			btc := getPriceAPI("BTC",strconv.Itoa(timeStamp))
			return fmt.Sprint(v*btc)
		}
	} else if symbol == "NEO" || symbol == "bNEO"{
		if v,ok := neoPrice[timeStamp];!ok{
			for j := 0; j < 999; j++ {
				if n,yes := neoPrice[timeStamp+1+j] ; yes {
					return fmt.Sprint(n)

				}
			}
			for j := 0; j < -999; j-- {
				if n,yes := neoPrice[timeStamp-1+j] ; yes {
					return fmt.Sprint(n)

				}
			}
		} else {
			return fmt.Sprint(v)
		}
	} else if symbol == "pONT" {
		if v,ok := ontPrice[timeStamp];!ok{
			for j := 0; j < 999; j++ {
				if n,yes := ontPrice[timeStamp+1+j] ; yes {
					return fmt.Sprint(n)

				}
			}
			for j := 0; j < -999; j-- {
				if n,yes := ontPrice[timeStamp-1+j] ; yes {
					return fmt.Sprint(n)

				}
			}
		} else {
			return fmt.Sprint(v)
		}
	} else if symbol == "FLM" {
		if v,ok := flmPrice[timeStamp];!ok{
			for j := 0; j < 999; j++ {
				if n,yes := flmPrice[timeStamp+1+j] ; yes {
					return fmt.Sprint(n)

				}
			}
			for j := 0; j < -999; j-- {
				if n,yes := flmPrice[timeStamp-1+j] ; yes {
					return fmt.Sprint(n)

				}
			}
		} else {
			return fmt.Sprint(v)
		}
	} else if symbol == "fUSDT" {
		return  fmt.Sprint(1)
	} else if symbol == "fWETH" {
		price := getPriceAPI("ETH",strconv.Itoa(timeStamp))
		return fmt.Sprint(price)
	} else if symbol == "fWBTC" {
		price := getPriceAPI("BTC",strconv.Itoa(timeStamp))
		return fmt.Sprint(price)
	}
	return ""
}

func (txs *Txs) ToCSV(w io.Writer) error {
	n := csv.NewWriter(w)
	err := n.Write([]string{"TxHash", "TimeStamp","Height","Address","Token0","Token0Price","Token1","Token1Price","Token2","Token2Price","Token3","Token3Price","SysFee","NetFee"})
	if err != nil {
		return err
	}
	for _, tx := range *txs {
		err := n.Write([]string{tx.TxHash, tx.TimeStamp, tx.Height,tx.Address,tx.Token0,tx.Token0Price,tx.Token1,tx.Token1Price,tx.Token2,tx.Token2Price,tx.Token3,tx.Token3Price,tx.SysFee,tx.NetFee})
		if err != nil {
			return err
		}
	}

	n.Flush()
	return n.Error()
}

func readCSV(fileName string) map[int]float64 {
	csvFile, err := os.Open(fileName)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	defer csvFile.Close()

	// Parse the file
	reader := csv.NewReader(bufio.NewReader(csvFile))
	var prices = make(map[int]float64)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		timeStamp,err := strconv.Atoi(line[4][0:10])
		if err != nil {
			log.Fatal(err)
		}
		price, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			log.Fatal(err)
		}

		prices[timeStamp] = price


	}
	return prices


}
