// Generate Tag data sets
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/iomz/go-llrp/binutil"
	"github.com/iomz/gosstrak/scheme"
	"gopkg.in/alecthomas/kingpin.v2"
)

type UIIParam struct {
	Type                  string
	Scheme                string
	CompanyPrefix         string
	ItemReference         string
	AssetType             string
	OwnerCode             string
	DataIdentifier        string
	IssuingAgencyCode     string
	CompanyIdentification string
	ExtDigits             int
	IARMaxDigits          int
	NoFilter              bool
}

var (
	// kingpin app
	app       = kingpin.New("gendataset", "A tool to generate dataset for simulation.")
	matchPct  = app.Flag("match", "match pct").Default("100").Int()
	nSub      = app.Flag("nsub", "nSub.").Default("1000").Int()
	NumSerial = 100
)

func getTargetDir() string {
	return os.Getenv("GOPATH") + fmt.Sprintf("/src/github.com/iomz/gosstrak/test/data/simulation/dataset%v-%vpct", *nSub, *matchPct)
}

func filterWriter(fs chan string) {
	f, err := os.OpenFile(path.Join(getTargetDir(), "filters.csv"), os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for {
		select {
		case s := <-fs:
			if len(s) != 0 {
				w.WriteString(s + "\n")
				w.Flush()
			}
		}
	}
}

func generateNonZeroDigits(len int) (s string) {
	ok := false
	for !ok {
		s = binutil.GenerateNLengthDigitString(len)
		if !strings.HasPrefix(s, "0") {
			ok = true
		}
	}
	return s
}

func generateUIISet(wg *sync.WaitGroup, q chan UIIParam, fq chan string) {
	defer wg.Done()
	for {
		param, ok := <-q
		if !ok {
			return
		}
		// do stuff
		switch param.Scheme {
		case "sgtin-96":
			schemeDir := path.Join(getTargetDir(), param.Scheme, param.CompanyPrefix)
			os.MkdirAll(schemeDir, 0755)
			fileName := path.Join(schemeDir, param.ItemReference)
			if _, err := os.Stat(fileName); !os.IsNotExist(err) {
				break
			}
			f, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}

			if !param.NoFilter {
				bs, opt := scheme.MakeEPC(true, param.Scheme, "3", param.CompanyPrefix, "", "", "", "", "")
				fq <- scheme.PrintID(bs, opt)
				bs, opt = scheme.MakeEPC(true, param.Scheme, "3", param.CompanyPrefix, param.ItemReference, "", "", "", "")
				fq <- scheme.PrintID(bs, opt)
			}

			w := bufio.NewWriter(f)
			for ser := 0; ser < NumSerial; ser++ {
				bs, opt := scheme.MakeEPC(false, param.Scheme, "3", param.CompanyPrefix, param.ItemReference, "", strconv.Itoa(ser), "", "")
				w.WriteString(scheme.PrintID(bs, opt) + "\n")
			}
			w.Flush()
			f.Close()
		case "sscc-96":
			schemeDir := path.Join(getTargetDir(), param.Scheme)
			os.MkdirAll(schemeDir, 0755)
			fileName := path.Join(schemeDir, param.CompanyPrefix)
			if _, err := os.Stat(fileName); !os.IsNotExist(err) {
				break
			}
			f, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}

			if !param.NoFilter {
				bs, opt := scheme.MakeEPC(true, param.Scheme, "3", param.CompanyPrefix, "", "", "", "", "")
				fq <- scheme.PrintID(bs, opt)
			}

			w := bufio.NewWriter(f)
			for ext := 0; ext < NumSerial; ext++ {
				bs, opt := scheme.MakeEPC(false, param.Scheme, "3", param.CompanyPrefix, "", generateNonZeroDigits(param.ExtDigits), "", "", "")
				w.WriteString(scheme.PrintID(bs, opt) + "\n")
			}
			w.Flush()
			f.Close()
		case "giai-96":
			schemeDir := path.Join(getTargetDir(), param.Scheme)
			os.MkdirAll(schemeDir, 0755)
			fileName := path.Join(schemeDir, param.CompanyPrefix)
			if _, err := os.Stat(fileName); !os.IsNotExist(err) {
				break
			}
			f, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}

			if !param.NoFilter {
				bs, opt := scheme.MakeEPC(true, param.Scheme, "3", param.CompanyPrefix, "", "", "", "", "")
				fq <- scheme.PrintID(bs, opt)
			}

			w := bufio.NewWriter(f)
			for iar := 0; iar < NumSerial; iar++ {
				iarLen := binutil.GenerateRandomInt(1, param.IARMaxDigits)
				bs, opt := scheme.MakeEPC(false, param.Scheme, "3", param.CompanyPrefix, "", "", "", generateNonZeroDigits(iarLen), "")
				w.WriteString(scheme.PrintID(bs, opt) + "\n")
			}
			w.Flush()
			f.Close()
		case "grai-96":
			schemeDir := path.Join(getTargetDir(), param.Scheme, param.CompanyPrefix)
			os.MkdirAll(schemeDir, 0755)
			fileName := path.Join(schemeDir, param.AssetType)
			if _, err := os.Stat(fileName); !os.IsNotExist(err) {
				log.Fatal(err)
			}
			f, err := os.Create(fileName)
			if err != nil {
				log.Fatal(err)
			}

			if !param.NoFilter {
				bs, opt := scheme.MakeEPC(true, param.Scheme, "3", param.CompanyPrefix, "", "", "", "", "")
				fq <- scheme.PrintID(bs, opt)
				bs, opt = scheme.MakeEPC(true, param.Scheme, "3", param.CompanyPrefix, "", "", "", "", param.AssetType)
				fq <- scheme.PrintID(bs, opt)
			}

			w := bufio.NewWriter(f)
			for ser := 0; ser < NumSerial; ser++ {
				bs, opt := scheme.MakeEPC(false, param.Scheme, "3", param.CompanyPrefix, "", "", strconv.Itoa(ser), "", param.AssetType)
				w.WriteString(scheme.PrintID(bs, opt) + "\n")
			}
			w.Flush()
			f.Close()
		case "17363":
			schemeDir := path.Join(getTargetDir(), param.Type+param.Scheme)
			os.MkdirAll(schemeDir, 0755)
			fileName := path.Join(schemeDir, param.OwnerCode)
			if _, err := os.Stat(fileName); !os.IsNotExist(err) {
				break
			}
			f, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}

			if !param.NoFilter {
				bs, opt := scheme.MakeISO(true, param.Scheme, param.OwnerCode, "", "", "", "", "", "")
				fq <- scheme.PrintID(bs, opt)
			}

			w := bufio.NewWriter(f)
			for ser := 0; ser < NumSerial; ser++ {
				bs, opt := scheme.MakeISO(false, param.Scheme, param.OwnerCode, "", strconv.Itoa(ser), "", "", "", "")
				w.WriteString(scheme.PrintID(bs, opt) + "\n")
			}
			w.Flush()
			f.Close()
		case "17365":
			schemeDir := path.Join(getTargetDir(), param.Type+param.Scheme)
			os.MkdirAll(schemeDir, 0755)
			fileName := path.Join(schemeDir, param.CompanyIdentification)
			if _, err := os.Stat(fileName); !os.IsNotExist(err) {
				break
			}
			f, err := os.Create(fileName)
			if err != nil {
				panic(err)
			}

			if !param.NoFilter {
				bs, opt := scheme.MakeISO(true, param.Scheme, "", "", "", param.DataIdentifier, param.IssuingAgencyCode, "", "")
				fq <- scheme.PrintID(bs, opt)
				bs, opt = scheme.MakeISO(true, param.Scheme, "", "", "", param.DataIdentifier, param.IssuingAgencyCode, param.CompanyIdentification, "")
				fq <- scheme.PrintID(bs, opt)
			}

			w := bufio.NewWriter(f)
			for ser := 0; ser < NumSerial; ser++ {
				serLen := binutil.GenerateRandomInt(10, 30)
				bs, opt := scheme.MakeISO(false, param.Scheme, "", "", "", param.DataIdentifier, param.IssuingAgencyCode, param.CompanyIdentification, binutil.GenerateNLengthAlphanumericString(serLen))
				w.WriteString(scheme.PrintID(bs, opt) + "\n")
			}
			w.Flush()
			f.Close()
		}
	}
}

func main() {
	parse := kingpin.MustParse(app.Parse(os.Args[1:]))
	_ = parse

	// create the target dir
	os.MkdirAll(getTargetDir(), 0755)
	log.Printf("Create target directory: %s", getTargetDir())

	// prepare the workers
	var wg sync.WaitGroup
	runtime.GOMAXPROCS(runtime.NumCPU())

	// prepare the filter writer
	fq := make(chan string)
	go filterWriter(fq)

	q := make(chan UIIParam, runtime.NumCPU())

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go generateUIISet(&wg, q, fq)
	}

	matchRepeat := int(float64(*matchPct) / 100 * float64(*nSub) / 20)
	mismatchRepeat := int(float64(100-*matchPct) / 100 * float64(*nSub) / 20)
	for i := 0; i < matchRepeat; i++ {
		var cpLen int
		log.Printf("Iteration: %v\n", i)
		cpLen = binutil.GenerateRandomInt(6, 12)
		cp := binutil.GenerateNLengthDigitString(cpLen)
		for j := 0; j < 10; j++ { // SGTIN x10 itemReference
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "sgtin-96",
				CompanyPrefix: cp,
				ItemReference: binutil.GenerateNLengthDigitString(13 - cpLen),
				NoFilter:      false,
			}
		} // 11 filters
		for j := 0; j < 2; j++ {
			cpLen = binutil.GenerateRandomInt(6, 12)
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "sscc-96",
				CompanyPrefix: binutil.GenerateNLengthDigitString(cpLen),
				ExtDigits:     17 - cpLen,
				NoFilter:      false,
			}
		} // 2 filters
		for j := 0; j < 2; j++ {
			cpLen = binutil.GenerateRandomInt(6, 12)
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "giai-96",
				CompanyPrefix: binutil.GenerateNLengthDigitString(cpLen),
				IARMaxDigits:  25 - cpLen,
				NoFilter:      false,
			}
		} // 2 filters
		cpLen = binutil.GenerateRandomInt(6, 11)
		cp = binutil.GenerateNLengthDigitString(cpLen)
		for j := 0; j < 2; j++ { // GRAI x2 assetType
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "grai-96",
				CompanyPrefix: cp,
				AssetType:     binutil.GenerateNLengthDigitString(12 - cpLen),
				NoFilter:      false,
			}
		} // 3 filters
		q <- UIIParam{
			Type:      "iso",
			Scheme:    "17363",
			OwnerCode: binutil.GenerateNLengthAlphabetString(3),
			NoFilter:  false,
		} // 1 filter
		ciLen := binutil.GenerateRandomInt(3, 7)
		q <- UIIParam{
			Type:                  "iso",
			Scheme:                "17365",
			DataIdentifier:        "25S",
			IssuingAgencyCode:     "U",
			CompanyIdentification: binutil.GenerateNLengthAlphanumericString(ciLen),
			NoFilter:              false,
		} // 1 filter
	}
	for i := 0; i < mismatchRepeat; i++ {
		var cpLen int
		log.Printf("Iteration: %v\n", i)
		cpLen = binutil.GenerateRandomInt(6, 12)
		cp := binutil.GenerateNLengthDigitString(cpLen)
		for j := 0; j < 10; j++ { // SGTIN x10 itemReference
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "sgtin-96",
				CompanyPrefix: cp,
				ItemReference: binutil.GenerateNLengthDigitString(13 - cpLen),
				NoFilter:      true,
			}
		} // 11 filters
		for j := 0; j < 2; j++ {
			cpLen = binutil.GenerateRandomInt(6, 12)
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "sscc-96",
				CompanyPrefix: binutil.GenerateNLengthDigitString(cpLen),
				ExtDigits:     17 - cpLen,
				NoFilter:      true,
			}
		} // 2 filters
		for j := 0; j < 2; j++ {
			cpLen = binutil.GenerateRandomInt(6, 12)
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "giai-96",
				CompanyPrefix: binutil.GenerateNLengthDigitString(cpLen),
				IARMaxDigits:  25 - cpLen,
				NoFilter:      true,
			}
		} // 2 filters
		cpLen = binutil.GenerateRandomInt(6, 11)
		cp = binutil.GenerateNLengthDigitString(cpLen)
		for j := 0; j < 2; j++ { // GRAI x2 assetType
			q <- UIIParam{
				Type:          "epc",
				Scheme:        "grai-96",
				CompanyPrefix: cp,
				AssetType:     binutil.GenerateNLengthDigitString(12 - cpLen),
				NoFilter:      true,
			}
		} // 3 filters
		q <- UIIParam{
			Type:      "iso",
			Scheme:    "17363",
			OwnerCode: binutil.GenerateNLengthAlphabetString(3),
			NoFilter:  true,
		} // 1 filter
		ciLen := binutil.GenerateRandomInt(3, 7)
		q <- UIIParam{
			Type:                  "iso",
			Scheme:                "17365",
			DataIdentifier:        "25S",
			IssuingAgencyCode:     "U",
			CompanyIdentification: binutil.GenerateNLengthAlphanumericString(ciLen),
			NoFilter:              true,
		} // 1 filter
	}

	close(q)

	wg.Wait()
}
