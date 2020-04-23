package main

import (
	"bufio"
	"encoding/json"
	//	"flag"
	"fmt"
	//	"io"
	//	"io/ioutil"
	"log"
	//	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	//	"time"
)

//структура для хранения результатов
type words struct {
	sync.Mutex //добавить в структуру мьютекс
	found      map[string]string
}

//Инициализация области памяти
func newWords() *words {
	return &words{found: map[string]string{}}
}

//Фиксируем вхождение слова
func (w *words) add(word string, WS string) {
	w.Lock()         //Заблокировать объект
	defer w.Unlock() // По завершению, разблокировать
	WorkStatus, ok := w.found[word]
	if !ok { //т.е. если ID запроса не найдено заводим новый элемент слайса
		w.found[word] = WS
		return
	}
	// слово найдено в очередной раз , увеличим счетчик у элемента слайса
	w.found[word] = WorkStatus + "," + WS
}

//
//XGigabitEthernet0/0/3 transceiver information:
//-------------------------------------------------------------
//Common information:
//  Transceiver Type               :1X_CopperPassive_SFP
//  Connector Type                 :Copper Pigtail
//  Wavelength(nm)                 :-
//  Transfer Distance(m)           :1(copper)
//  Digital Diagnostic Monitoring  :NO
//  Vendor Name                    :TIMEINTERCONNECT
//  Vendor Part Number             :D09181-4A
//  Ordering Name                  :
//-------------------------------------------------------------
//Manufacture information:
//  Manu. Serial Number            :000405
//  Manufacturing Date             :2018-08-13
//  Vendor Name                    :TIMEINTERCONNECT
//-------------------------------------------------------------
//Info: Port XGigabitEthernet1/0/2, transceiver is absent.
//
//
//XGigabitEthernet1/0/3 transceiver information:
//-------------------------------------------------------------
//Common information:
//  ---- More ----                                            Transceiver Type               :1X_CopperPassive_SFP
//  Connector Type                 :Copper Pigtail
//  Wavelength(nm)                 :-
//  Transfer Distance(m)           :1(copper)
//  Digital Diagnostic Monitoring  :NO
//  Vendor Name                    :TIMEINTERCONNECT
//  Vendor Part Number             :D09181-4A
//  Ordering Name                  :
//-------------------------------------------------------------
//Manufacture information:
//  Manu. Serial Number            :000443
//  Manufacturing Date             :2018-08-13
//  Vendor Name                    :TIMEINTERCONNECT

func main() {
	//Создание структуры хранения результатов
	w := newWords()

	floger, err := os.OpenFile("huawei_audit.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer floger.Close()

	log.SetOutput(floger)

	fname := "5720-trans.txt"
	id := "m39-3-1-mdp-zsw-01"
	parse5720(id, fname, w)

	fname = "CE6851HI2-trans.txt"
	id = "m39-3-1-mdp-dsw-01"
	parse5720(id, fname, w)

	var keys []string
	for k, _ := range w.found {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// To perform the opertion you want
	fmt.Printf("{\n\"data\": [\n")
	for _, k := range keys {
		//! fmt.Println("Key:", k, "Value:", w.found[k])
		var dat map[string]interface{}
		err := json.Unmarshal([]byte("{ "+w.found[k]+" }"), &dat)
		if err != nil {
			//!fmt.Printf("ErrorUnmarshal id = %d \t%s\n", k, "{ " + w.found[k] + " }")
			continue
		} else {
			fmt.Printf("{ "+"\"Name\": \"%s\",%s\n", k, w.found[k]+" },")

		}
	}
	fmt.Printf("]\n}")

}

func parse5720(id string, fsrc string, dict *words) {
	f, err := os.Open(fsrc)
	if err != nil {
		log.Printf("Error open %s \n", fsrc)
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	log.Println(" ")
	idkey := ""
	line := ""
	Manu := 0 //логический признак Manufacture information:

	for scanner.Scan() {
		line = strings.TrimSpace(fmt.Sprintf("%s", scanner.Text()))
		if strings.Contains(line, "transceiver information:") { // если найдено начало структуры
			Manu = 0 //сброс флага
			k := strings.Split(line, " ")
			idkey = id + ": " + k[0]
			log.Printf("found %s \n", idkey)
			continue
		}

		if strings.Contains(line, "Transceiver Type") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Transceiver Type") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Connector Type") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Connector Type") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Wavelength") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Wavelength") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Transfer Distance") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Transfer Distance") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Digital Diagnostic Monitoring") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Digital Diagnostic Monitoring") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Vendor Name") && Manu == 0 { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Vendor Name") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Vendor Part Number") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Vendor Part Number") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Ordering Name") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Ordering Name") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Manufacture information:") {
			Manu = 1
		}

		if strings.Contains(line, "Manu. Serial Number") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Manu. Serial Number") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Manufacturing Date") { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Manufacturing Date") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)
			dict.add(idkey, vS)
			continue
		}

		if strings.Contains(line, "Vendor Name") && Manu == 1 { // если найден элемент структуры
			k := strings.Split(line, ":")
			vS := jsElement("Manu Vendor Name") + fmt.Sprintf("%s", strings.TrimSpace(k[1])) + "\""
			log.Printf("found  %s \n", vS)

			dict.add(idkey, vS) //разобрали последний элемент структуры, записываем данные в массив

			Manu = 0   //сброс флага
			idkey = "" //сброс клуча
			continue
		}

		//                vS = jsElement(idkey) + fmt.Sprintf("%s", strings.TrimSpace(dat)) + "\""
		//Добавим результат выполнения запроса Ответ сервера
		//		dict.add(idkey, vS)

	}
	return
}

func jsElement(key string) string {
	if "Transceiver Type" == key {
		return "\"TransceiverType\": \""
	} else if "Connector Type" == key {
		return "\"ConnectorType\": \""
	} else if "Wavelength" == key {
		return "\"Wavelength\": \""
	} else if "Transfer Distance" == key {
		return "\"TransferDistance\": \""
	} else if "Digital Diagnostic Monitoring" == key {
		return "\"DigitalDiagnosticMonitoring\": \""
	} else if "Vendor Name" == key {
		return "\"VendorName\": \""
	} else if "Vendor Part Number" == key {
		return "\"VendorPartNumber\": \""
	} else if "Ordering Name" == key {
		return "\"OrderingName\": \""
	} else if "Manu. Serial Number" == key {
		return "\"ManuSerialNumber\": \""
	} else if "Manufacturing Date" == key {
		return "\"ManufacturingDate\": \""
	} else if "Manu Vendor Name" == key {
		return "\"ManuVendorName\": \""
	}
	return "0"
}
