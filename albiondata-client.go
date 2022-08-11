package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"github.com/broderickhyman/albiondata-client/client"
	"github.com/broderickhyman/albiondata-client/log"
	"github.com/broderickhyman/albiondata-client/systray"
	"github.com/go-sql-driver/mysql"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"
)

var version string

func init() {
	client.ConfigGlobal.SetupFlags()
}

type Categories struct {
	Categories []Category `yaml:"Categories"`
}

type Category struct {
	CategoryName string `yaml:"Category"`
	SubItems     []Item `yaml:"Items"`
}

type Item struct {
	Name     string    `yaml:"Name"`
	CodeName string    `yaml:"CodeName"`
	Toggle   bool      `yaml:"Toggle"`
	Weight   []float32 `yaml:"Weight"`
}

type dataItem struct {
	KeyId            int
	TradeId          string
	ItemTypeId       string
	ItemGroupTypeId  string
	LocationId       int
	QualityLevel     int
	EnchantmentLevel int
	UnitPriceSilver  int
	Amount           int
	AuctionType      string
	Expires          string
	LastUpdated      string
}

type tableCell struct {
	Text        string
	LastUpdated string
	Color       color.NRGBA
}

var db *sql.DB

//The lower the value, the greener the margin %s
//The higher the value, the redder the margin %s
//I like about 1500
const marginConst float64 = 1500

//The lower the value, the green the PPT
//The higher the value, the redder the PPT
//I like about 50000000
const tripConst float64 = 50000000

const ItemNameColumn int = 0
const WeightColumn int = 13
const MPPTColumn int = 14
const MPMColumn int = 15
const IPPTColumn int = 16
const IPMColumn int = 17

func main() {
	//startUpdater()

	// Setup database connection
	// Capture connection properties.
	cfg := mysql.Config{
		User:   os.Getenv("DBUSER"),
		Passwd: os.Getenv("DBPASS"),
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "albiontradelogs",
	}

	fmt.Println("Initializing connection!:")
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	go systray.Run()
	c := client.NewClient(version)
	go c.Run()
	/*if err != nil {
		log.Error(err)
		log.Error("The program encountered an error. Press any key to close this window.")
		var b = make([]byte, 1)
		_, _ = os.Stdin.Read(b)
	}*/

	guiInit()
}

/*func startUpdater() {
	if version != "" && !strings.Contains(version, "dev") {
		u := updater.NewUpdater(
			version,
			"broderickhyman",
			"albiondata-client",
			"update-",
		)

		go func() {
			maxTries := 2
			for i := 0; i < maxTries; i++ {
				err := u.BackgroundUpdater()
				if err != nil {
					log.Error(err.Error())
					log.Info("Will try again in 60 seconds. You may need to run the client as Administrator.")
					// Sleep and hope the network connects
					time.Sleep(time.Second * 60)
				} else {
					break
				}
			}
		}()
	}
}*/

func guiInit() {
	a := app.New()
	w := a.NewWindow("Albion Trading Data")
	path := "itemData.yaml"
	var file, err = os.ReadFile(path)
	var itemDatas Categories
	if err != nil {
		log.Info("Error opening file")
	} else {
		err2 := yaml.Unmarshal(file, &itemDatas)
		if err2 != nil {
			log.Info("Error unmarhsalling: %s", err2)
		}
		log.Info("Info2: ", itemDatas)
	}

	towns := []string{"Bridgewatch", "Lymhurst", "Fort Sterling", "Thetford", "Martlock", "Caerleon"}
	townIDs := []string{"2004", "1002", "4002", "0007", "3008", "3005"}

	//Search bar that filters items
	searchBar := widget.NewEntry()

	//Print out the top row with the towns
	var topRow = []string{}
	topRow = append(topRow, "Items")
	for _, town := range towns {
		topRow = append(topRow, town+" S")
		topRow = append(topRow, town+" B")
	}
	topRow = append(topRow, "Weight")
	topRow = append(topRow, "Max PPT")
	topRow = append(topRow, "Max PM")
	topRow = append(topRow, "Instant PPT")
	topRow = append(topRow, "Instant PM")
	topRow = append(topRow, " ")

	//width := len(topRow)
	width := 500
	tableData := make([][]tableCell, width)
	for i := range tableData {
		tableData[i] = make([]tableCell, width)
	}

	for i, _ := range topRow {
		tableData[0][i].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
		tableData[0][i].Text = topRow[i]
	}

	//Make table that holds our data
	tableHeader := widget.NewTable(
		func() (int, int) {
			return 1, len(tableData[0])
		},
		func() fyne.CanvasObject {
			return canvas.NewText("Hello world123", color.Black)
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*canvas.Text).Color = tableData[i.Row][i.Col].Color
			o.(*canvas.Text).Text = tableData[i.Row][i.Col].Text
		})

	table := widget.NewTable(
		func() (int, int) {
			return len(tableData), len(tableData[0])
		},
		func() fyne.CanvasObject {
			return canvas.NewText("Hello world123", color.Black)
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*canvas.Text).Color = tableData[i.Row][i.Col].Color
			o.(*canvas.Text).Text = tableData[i.Row][i.Col].Text
		})

	var values *fyne.Container
	var content *fyne.Container

	carryCapacity := "3626"
	carryCapacityBinding := binding.BindString(&carryCapacity)
	carryCapacityEntry := widget.NewEntryWithData(carryCapacityBinding)

	//Checkbox for Caerleon
	caerleonToggle := false
	caerleonToggleBind := binding.BindBool(&caerleonToggle)
	check := widget.NewCheckWithData("Include Caerleon", caerleonToggleBind)

	orderTypes := []string{"offer", "request"}
	confirm := widget.NewButton("Confirm", func() {
		currentIndex := 1

		//Grab the items that are toggled on
		for _, cats := range itemDatas.Categories {
			for _, items := range cats.SubItems {
				if items.Toggle == true {
					var queryRow dataItem
					foundAnItem := false
					//Check if the item has no tier
					result := db.QueryRow("SELECT * FROM traderecords WHERE itemTypeId = ?", items.CodeName)
					if err := result.Scan(&queryRow.KeyId, &queryRow.TradeId, &queryRow.ItemTypeId, &queryRow.ItemGroupTypeId, &queryRow.LocationId, &queryRow.QualityLevel, &queryRow.EnchantmentLevel, &queryRow.UnitPriceSilver, &queryRow.Amount, &queryRow.AuctionType, &queryRow.Expires, &queryRow.LastUpdated); err == nil {
						itemTableName := items.CodeName
						for i, town := range townIDs {
							for n, orderType := range orderTypes {
								result := db.QueryRow("SELECT * FROM traderecords WHERE ItemTypeId = ? AND LocationId = ? AND AuctionType = ?", itemTableName, town, orderType)
								tierErr := result.Scan(&queryRow.KeyId, &queryRow.TradeId, &queryRow.ItemTypeId, &queryRow.ItemGroupTypeId, &queryRow.LocationId, &queryRow.QualityLevel, &queryRow.EnchantmentLevel, &queryRow.UnitPriceSilver, &queryRow.Amount, &queryRow.AuctionType, &queryRow.Expires, &queryRow.LastUpdated)
								if tierErr == nil {
									//Remove the unnecessary 0's and add commas
									p := message.NewPrinter(language.English)
									word := p.Sprintf("%d", queryRow.UnitPriceSilver/10000)
									tableData[currentIndex][(i*2)+n+1].Text = word
									tableData[currentIndex][(i*2)+n+1].LastUpdated = queryRow.LastUpdated

									timeVal, _ := time.Parse(time.RFC3339Nano, tableData[currentIndex][(i*2)+n+1].LastUpdated)
									timeSub := time.Now().Sub(timeVal)
									if timeSub.Hours() <= 4 {
										tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
									} else if timeSub.Hours() <= 48 {
										//Scale between 100-255
										//If the cell is 4 hours old, it's 255 (max)
										//If the cell is 48 hours old, it's 100 (min that's still fairly legible)
										datedness := uint8((255 - (155 * ((timeSub.Hours() - 4) / 44))))
										tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: datedness}
									} else {
										tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 100}
									}

									foundAnItem = true
								} else {
									//Search error, leave the cell blank and make the color transparent
									tableData[currentIndex][(i*2)+n+1].Text = "-"
									tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 0}
								}
							}
						}

						if foundAnItem == true {
							//Put in the item name
							tableData[currentIndex][ItemNameColumn].Text = items.Name
							tableData[currentIndex][ItemNameColumn].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

							//Put in the weight
							itemWeight := strconv.FormatFloat(float64(items.Weight[0]), 'f', 1, 32)
							tableData[currentIndex][WeightColumn].Text = itemWeight
							tableData[currentIndex][WeightColumn].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

							//Next item
							currentIndex += 1
						}
					} else {
						//Otherwise, Check if the item has tiers
						if err == sql.ErrNoRows {
							foundAnItem = false
							//Todo: Some items have a tier 1, so this should be 8 for some
							for i := 0; i < 7; i++ {
								itemTableName := "T" + strconv.Itoa(i+1) + "_" + items.CodeName
								for i, town := range townIDs {
									for n, orderType := range orderTypes {
										result := db.QueryRow("SELECT * FROM traderecords WHERE ItemTypeId = ? AND LocationId = ? AND AuctionType = ?", itemTableName, town, orderType)
										tierErr := result.Scan(&queryRow.KeyId, &queryRow.TradeId, &queryRow.ItemTypeId, &queryRow.ItemGroupTypeId, &queryRow.LocationId, &queryRow.QualityLevel, &queryRow.EnchantmentLevel, &queryRow.UnitPriceSilver, &queryRow.Amount, &queryRow.AuctionType, &queryRow.Expires, &queryRow.LastUpdated)
										if tierErr == nil {
											//Remove the unnecessary 0's and add commas
											p := message.NewPrinter(language.English)
											word := p.Sprintf("%d", queryRow.UnitPriceSilver/10000)
											tableData[currentIndex][(i*2)+n+1].Text = word
											tableData[currentIndex][(i*2)+n+1].LastUpdated = queryRow.LastUpdated

											timeVal, _ := time.Parse(time.RFC3339Nano, tableData[currentIndex][(i*2)+n+1].LastUpdated)
											timeSub := time.Now().Sub(timeVal)
											if timeSub.Hours() <= 4 {
												tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
											} else if timeSub.Hours() <= 48 {
												//Scale between 100-255
												//If the cell is 4 hours old, it's 255 (max)
												//If the cell is 48 hours old, it's 100 (min that's still fairly legible)
												datedness := uint8((255 - (155 * ((timeSub.Hours() - 4) / 44))))
												tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: datedness}
											} else {
												tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 100}
											}

											foundAnItem = true
										} else {
											tableData[currentIndex][(i*2)+n+1].Text = "-"
											tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
										}
									}
								}

								if foundAnItem == true {
									//Put in the item name
									tableData[currentIndex][ItemNameColumn].Text = "T" + strconv.Itoa(i+1) + " " + items.Name
									tableData[currentIndex][ItemNameColumn].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

									//Put in the weight
									itemWeight := strconv.FormatFloat(float64(items.Weight[i]), 'f', 1, 32)
									tableData[currentIndex][WeightColumn].Text = itemWeight
									tableData[currentIndex][WeightColumn].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

									//Next item
									currentIndex += 1
								}
								//Check if the item has enchantments
								for n := 0; n < 3; n++ {
									foundAnItem = false
									itemTableName2 := itemTableName + "_LEVEL" + strconv.Itoa(n+1) + "@" + strconv.Itoa(n+1)

									for i, town := range townIDs {
										for n, orderType := range orderTypes {
											result := db.QueryRow("SELECT * FROM traderecords WHERE ItemTypeId = ? AND LocationId = ? AND AuctionType = ?", itemTableName2, town, orderType)
											tierErr := result.Scan(&queryRow.KeyId, &queryRow.TradeId, &queryRow.ItemTypeId, &queryRow.ItemGroupTypeId, &queryRow.LocationId, &queryRow.QualityLevel, &queryRow.EnchantmentLevel, &queryRow.UnitPriceSilver, &queryRow.Amount, &queryRow.AuctionType, &queryRow.Expires, &queryRow.LastUpdated)
											if tierErr == nil {
												//Remove the unnecessary 0's and add commas
												p := message.NewPrinter(language.English)
												word := p.Sprintf("%d", queryRow.UnitPriceSilver/10000)
												tableData[currentIndex][(i*2)+n+1].Text = word
												tableData[currentIndex][(i*2)+n+1].LastUpdated = queryRow.LastUpdated

												timeVal, _ := time.Parse(time.RFC3339Nano, tableData[currentIndex][(i*2)+n+1].LastUpdated)
												timeSub := time.Now().Sub(timeVal)
												if timeSub.Hours() <= 4 {
													tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
												} else if timeSub.Hours() <= 48 {
													//Scale between 100-255
													//If the cell is 4 hours old, it's 255 (max)
													//If the cell is 48 hours old, it's 100 (min that's still fairly legible)
													datedness := uint8((255 - (155 * ((timeSub.Hours() - 4) / 44))))
													tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: datedness}
												} else {
													tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 100}
												}

												foundAnItem = true
											} else {
												tableData[currentIndex][(i*2)+n+1].Text = "-"
												tableData[currentIndex][(i*2)+n+1].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
											}
										}
									}

									if foundAnItem == true {
										//Put in the item name
										tableData[currentIndex][ItemNameColumn].Text = "T" + strconv.Itoa(i+1) + "." + strconv.Itoa(n+1) + " " + items.Name
										tableData[currentIndex][ItemNameColumn].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

										//Put in the weight
										itemWeight := strconv.FormatFloat(float64(items.Weight[i]), 'f', 1, 32)
										tableData[currentIndex][WeightColumn].Text = itemWeight
										tableData[currentIndex][WeightColumn].Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

										//Next item
										currentIndex += 1
									}
								}
							}
						}
					}
				}
			}
		}

		//Now that we have all of our sell/buy order data
		//We iterate through the table a second time and fill
		//out our calculations
		for secondPassIndex := 1; secondPassIndex < currentIndex; secondPassIndex++ {
			//We're going to get the highest/lowest buy/sell orders
			//First, Initialize the values
			//TODO: We kinda assume the first town is filled out, this might bug out if it's not
			lowestSell := tableData[secondPassIndex][1].Text
			highestSell := tableData[secondPassIndex][1].Text
			lowestBuy := tableData[secondPassIndex][2].Text
			highestBuy := tableData[secondPassIndex][2].Text
			lowestSellIndex := 1
			highestSellIndex := 1
			lowestBuyIndex := 2
			highestBuyIndex := 2

			//Next, Iterate through all the towns for our values
			for x, town := range towns {
				//Ignore Caerleon if it's not toggled on
				if (town == "Caerleon" && caerleonToggle == true) || town != "Caerleon" {
					//Remove commas and turn string to int
					lowestSellInt, err1 := strconv.Atoi(strings.Replace(lowestSell, ",", "", -1))
					highestSellInt, err1 := strconv.Atoi(strings.Replace(highestSell, ",", "", -1))
					lowestBuyInt, err1 := strconv.Atoi(strings.Replace(lowestBuy, ",", "", -1))
					highestBuyInt, err1 := strconv.Atoi(strings.Replace(highestBuy, ",", "", -1))
					sellInt, err2 := strconv.Atoi(strings.Replace(tableData[secondPassIndex][(x*2)+1].Text, ",", "", -1))
					buyInt, err2 := strconv.Atoi(strings.Replace(tableData[secondPassIndex][(x*2)+2].Text, ",", "", -1))

					if lowestSellInt > sellInt && err1 == nil && err2 == nil {
						lowestSell = tableData[secondPassIndex][(x*2)+1].Text
						lowestSellIndex = (x * 2) + 1
					}

					if highestSellInt < sellInt && err1 == nil && err2 == nil {
						highestSell = tableData[secondPassIndex][(x*2)+1].Text
						highestSellIndex = (x * 2) + 1
					}

					if lowestBuyInt > buyInt && err1 == nil && err2 == nil {
						lowestBuy = tableData[secondPassIndex][(x*2)+2].Text
						lowestBuyIndex = (x * 2) + 2
					}

					if highestBuyInt < buyInt && err1 == nil && err2 == nil {
						highestBuy = tableData[secondPassIndex][(x*2)+2].Text
						highestBuyIndex = (x * 2) + 2
					}
				}
			}

			//Highlight the lowest/highest buy/sell orders
			//Keep the old transparency
			tableData[secondPassIndex][lowestSellIndex].Color = color.NRGBA{R: 0, G: 127, B: 255, A: tableData[secondPassIndex][lowestSellIndex].Color.A}
			tableData[secondPassIndex][highestSellIndex].Color = color.NRGBA{R: 255, G: 255, B: 0, A: tableData[secondPassIndex][highestSellIndex].Color.A}
			tableData[secondPassIndex][lowestBuyIndex].Color = color.NRGBA{R: 0, G: 127, B: 255, A: tableData[secondPassIndex][lowestBuyIndex].Color.A}
			tableData[secondPassIndex][highestBuyIndex].Color = color.NRGBA{R: 255, G: 255, B: 0, A: tableData[secondPassIndex][highestBuyIndex].Color.A}
			//Turn everything into floats so we can calculate our stuffs
			lowestSellFloat, _ := strconv.ParseFloat(strings.Replace(lowestSell, ",", "", -1), 32)
			highestSellFloat, _ := strconv.ParseFloat(strings.Replace(highestSell, ",", "", -1), 32)
			lowestBuyFloat, _ := strconv.ParseFloat(strings.Replace(lowestBuy, ",", "", -1), 32)
			highestBuyFloat, _ := strconv.ParseFloat(strings.Replace(highestBuy, ",", "", -1), 32)
			carryCapacityFloat, _ := strconv.ParseFloat(carryCapacity, 32)
			itemWeightFloat, _ := strconv.ParseFloat(strings.Replace(tableData[secondPassIndex][WeightColumn].Text, ",", "", -1), 32)

			//Calculate the MPPT
			mpptVal := ((highestSellFloat - (highestSellFloat * 0.045)) - (lowestBuyFloat + (lowestBuyFloat * 0.015))) * (carryCapacityFloat / itemWeightFloat)
			tableData[secondPassIndex][MPPTColumn].Text = strconv.FormatFloat(mpptVal, 'f', 1, 64)
			//Give MPPT color
			if mpptVal <= 0 {
				tableData[secondPassIndex][MPPTColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
			} else {
				totalVal := 510 - (int(tripConst / mpptVal))

				if totalVal <= 0 {
					tableData[secondPassIndex][MPPTColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
				} else if totalVal > 0 && totalVal <= 255 {
					tableData[secondPassIndex][MPPTColumn].Color = color.NRGBA{R: 255, G: uint8(totalVal), B: 0, A: 255}
				} else if totalVal > 255 && totalVal <= 510 {
					tableData[secondPassIndex][MPPTColumn].Color = color.NRGBA{R: (255 - uint8(totalVal)), G: 255, B: 0, A: 255}
				} else {
					tableData[secondPassIndex][MPPTColumn].Color = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
				}
			}

			//Calculate the MPM
			mpmVal := (((highestSellFloat - (highestSellFloat * 0.045)) / (lowestBuyFloat + (lowestBuyFloat * 0.015))) - 1) * 100
			tableData[secondPassIndex][MPMColumn].Text = strconv.FormatFloat(mpmVal, 'f', 2, 64) + "%"
			//Give MPM color
			if mpmVal <= 0 {
				tableData[secondPassIndex][MPMColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
			} else {
				totalVal := 510 - (int(marginConst / mpmVal))

				if totalVal <= 0 {
					tableData[secondPassIndex][MPMColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
				} else if totalVal > 0 && totalVal <= 255 {
					tableData[secondPassIndex][MPMColumn].Color = color.NRGBA{R: 255, G: uint8(totalVal), B: 0, A: 255}
				} else if totalVal > 255 && totalVal <= 510 {
					tableData[secondPassIndex][MPMColumn].Color = color.NRGBA{R: (255 - uint8(totalVal)), G: 255, B: 0, A: 255}
				} else {
					tableData[secondPassIndex][MPMColumn].Color = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
				}
			}

			//Calculate the IPPT
			ipptVal := (highestBuyFloat - (lowestSellFloat + (lowestSellFloat * 0.03))) * (carryCapacityFloat / itemWeightFloat)
			tableData[secondPassIndex][IPPTColumn].Text = strconv.FormatFloat(ipptVal, 'f', 1, 64)
			//Give IPPT color
			if ipptVal <= 0 {
				tableData[secondPassIndex][IPPTColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
			} else {
				totalVal := 510 - (int(tripConst / ipptVal))

				if totalVal <= 0 {
					tableData[secondPassIndex][IPPTColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
				} else if totalVal > 0 && totalVal <= 255 {
					tableData[secondPassIndex][IPPTColumn].Color = color.NRGBA{R: 255, G: uint8(totalVal), B: 0, A: 255}
				} else if totalVal > 255 && totalVal <= 510 {
					tableData[secondPassIndex][IPPTColumn].Color = color.NRGBA{R: (255 - uint8(totalVal)), G: 255, B: 0, A: 255}
				} else {
					tableData[secondPassIndex][IPPTColumn].Color = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
				}
			}

			//Calculate the IPM
			ipmVal := ((highestBuyFloat / (lowestSellFloat + (lowestSellFloat * 0.03))) - 1) * 100
			tableData[secondPassIndex][IPMColumn].Text = strconv.FormatFloat(ipmVal, 'f', 2, 64) + "%"
			//Give IPM color
			if ipmVal <= 0 {
				tableData[secondPassIndex][IPMColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
			} else {
				totalVal := 510 - (int(marginConst / ipmVal))

				if totalVal <= 0 {
					tableData[secondPassIndex][IPMColumn].Color = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
				} else if totalVal > 0 && totalVal <= 255 {
					tableData[secondPassIndex][IPMColumn].Color = color.NRGBA{R: 255, G: uint8(totalVal), B: 0, A: 255}
				} else if totalVal > 255 && totalVal <= 510 {
					tableData[secondPassIndex][IPMColumn].Color = color.NRGBA{R: (255 - uint8(totalVal)), G: 255, B: 0, A: 255}
				} else {
					tableData[secondPassIndex][IPMColumn].Color = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
				}
			}
		}

		tableHeader := widget.NewTable(
			func() (int, int) {
				return 1, len(tableData[0])
			},
			func() fyne.CanvasObject {
				return canvas.NewText("Hello world123", color.Black)
			},
			func(i widget.TableCellID, o fyne.CanvasObject) {
				o.(*canvas.Text).Color = tableData[i.Row][i.Col].Color
				o.(*canvas.Text).Text = tableData[i.Row][i.Col].Text
			})

		//Now display the data
		table = widget.NewTable(
			func() (int, int) {
				return len(tableData) - 1, len(tableData[0]) - 1
			},
			func() fyne.CanvasObject {
				return canvas.NewText("Hello world123", color.Black)
			},
			func(i widget.TableCellID, o fyne.CanvasObject) {
				//Ignore the header since we have tableHeader
				if i.Row != len(tableData) {
					o.(*canvas.Text).Color = tableData[i.Row+1][i.Col].Color
					o.(*canvas.Text).Text = tableData[i.Row+1][i.Col].Text
				}
			})

		combinedTable := container.NewBorder(
			tableHeader,
			nil,
			nil,
			nil,
			table,
		)
		values = container.NewBorder(
			content,
			nil,
			nil,
			nil,
			combinedTable,
		)

		w.SetContent(values)
	})

	//Accordion for displaying items, which gets shoved under titleAccordion
	catAccordion := widget.NewAccordion()
	for i, category1 := range itemDatas.Categories {
		var theGroupToDo []string
		individualCheck := []fyne.CanvasObject{}
		for n, items := range category1.SubItems {
			boolThing := binding.BindBool(&itemDatas.Categories[i].SubItems[n].Toggle)
			individualCheck = append(individualCheck, widget.NewCheckWithData(items.Name, boolThing))
			theGroupToDo = append(theGroupToDo, items.Name)
		}
		individualCheck = append(individualCheck, widget.NewSeparator())
		container1 := container.NewVBox(
			individualCheck...,
		)
		newAccordionItem := widget.NewAccordionItem(category1.CategoryName, container1)
		catAccordion.Append(newAccordionItem)
	}
	//Accordion for displaying categories
	titleAccordion := widget.NewAccordion()
	titleAccordionItem := widget.NewAccordionItem("Categories", catAccordion)
	titleAccordion.Append(titleAccordionItem)

	notes := canvas.NewText("Note: PPT = Profit Per Trip, PM = Profit Margin. Maximum assumes order setup fees (3%), Instant assumes no order setup fees. All assume premium tax (1.5%)", color.White)
	weightText := canvas.NewText("Carrying Capacity:", color.White)
	buttons := container.NewHBox(
		titleAccordion,
		container.NewGridWithColumns(
			2,
			weightText,
			carryCapacityEntry,
		),
		check,
		confirm,
	)
	content = container.NewVBox(
		searchBar,
		buttons,
		notes,
	)

	combinedTable := container.NewBorder(
		tableHeader,
		nil,
		nil,
		nil,
		table,
	)
	values = container.NewBorder(
		content,
		nil,
		nil,
		nil,
		combinedTable,
	)

	w.Resize(fyne.NewSize(2400, 1300))
	w.SetContent(values)
	w.ShowAndRun()
}
