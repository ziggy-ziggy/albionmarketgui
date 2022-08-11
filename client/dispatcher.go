package client

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"strings"

	"github.com/broderickhyman/albiondata-client/lib"
	"github.com/broderickhyman/albiondata-client/log"
	"gopkg.in/yaml.v2"

	"github.com/go-sql-driver/mysql"
)

type dispatcher struct {
	publicUploaders  []uploader
	privateUploaders []uploader
}

type MarketHistories struct {
	AlbionId        string          `yaml:"AlbionId"`
	LocationId      int             `yaml:"LocationId"`
	QualityLevel    string          `yaml:"QualityLevel"`
	Timescale       string          `yaml:"Timescale"`
	MarketHistories []MarketHistory `yaml:"MarketHistories"`
}
type MarketHistory struct {
	ItemAmount   string `yaml:"ItemAmount"`
	SilverAmount string `yaml:"SilverAmount"`
	Timestamp    string `yaml:"Timestamp"`
}

type Orders struct {
	Orders []Order `yaml:"Orders"`
}

//Order struct base
type Order struct {
	//EventType        string `json:"Orders"`
	Id               string `yaml:"Id"`
	ItemTypeId       string `yaml:"ItemTypeId"`
	ItemGroupTypeId  string `yaml:"ItemGroupTypeId"`
	LocationId       int    `yaml:"LocationId"`
	QualityLevel     string `yaml:"QualityLevel"`
	EnchantmentLevel string `yaml:"EnchantmentLevel"`
	UnitPriceSilver  string `yaml:"UnitPriceSilver"`
	Amount           string `yaml:"Amount"`
	AuctionType      string `yaml:"AuctionType"`
	Expires          string `yaml:"Expires"`
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

var (
	wsHub *WSHub
	dis   *dispatcher
)

var db *sql.DB

func createDispatcher() {
	dis = &dispatcher{
		publicUploaders:  createUploaders(strings.Split(ConfigGlobal.PublicIngestBaseUrls, ",")),
		privateUploaders: createUploaders(strings.Split(ConfigGlobal.PrivateIngestBaseUrls, ",")),
	}

	if ConfigGlobal.EnableWebsockets {
		wsHub = newHub()
		go wsHub.run()
		go runHTTPServer()
	}

	// Setup database connection
	// Capture connection properties.
	cfg := mysql.Config{
		//User:   os.Getenv("DBUSER"),
		User: "ziggy",
		//Passwd: os.Getenv("DBPASS"),
		Passwd: "applecaca123",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "albiontradelogs",
	}

	fmt.Println("Initializing connection!")
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

	//Make table if it's not already created
	result, err := db.Exec("CREATE TABLE IF NOT EXISTS traderecords (KeyId INT AUTO_INCREMENT, TradeId VARCHAR(255) NOT NULL, ItemTypeId VARCHAR(255) NOT NULL, ItemGroupTypeId VARCHAR(255) NOT NULL, LocationId INT NOT NULL, QualityLevel INT, EnchantmentLevel INT, UnitPriceSilver INT NOT NULL, Amount INT NOT NULL, AuctionType VARCHAR(128) NOT NULL, Expires VARCHAR(255) NOT NULL, LastUpdated VARCHAR(255) NOT NULL, PRIMARY KEY (KeyId), UNIQUE KEY(ItemTypeId, LocationId, EnchantmentLevel, AuctionType));")
	if err != nil {
		fmt.Errorf("creatingTable: %v", err)
	}
	_, err3 := result.LastInsertId()
	if err3 != nil {
		fmt.Errorf("creatingTable: %v", err3)
	}
	fmt.Println("Connected!")
}

func createUploaders(targets []string) []uploader {
	var uploaders []uploader
	for _, target := range targets {
		if target == "" {
			continue
		}
		if len(target) < 4 {
			log.Infof("Got an ingest target that was less than 4 characters, not a valid ingest target: %v", target)
			continue
		}

		if target[0:8] == "http+pow" {
			uploaders = append(uploaders, newHTTPUploaderPow(target))
		} else if target[0:4] == "http" {
			uploaders = append(uploaders, newHTTPUploader(target))
		} else if target[0:4] == "nats" {
			uploaders = append(uploaders, newNATSUploader(target))
		} else {
			log.Infof("An invalid ingest target was specified: %v", target)
		}
	}

	return uploaders
}

func sendMsgToPublicUploaders(upload interface{}, topic string, state *albionState) {
	data, err := json.Marshal(upload)
	if err != nil {
		log.Errorf("Error while marshalling payload for %v: %v", err, topic)
		return
	}
	//log.Infof("Data was: %s", data)

	// we initialize our Orders array
	var orders Orders
	var markethistories MarketHistories
	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'Orders' which we defined above
	yaml.Unmarshal(data, &orders)
	err2 := yaml.Unmarshal(data, &markethistories)
	if err2 != nil {
		log.Errorf("Error while marshalling payload for marketHistories: %v", err)
		return
	}
	fmt.Println("MarketHistory: ", markethistories)
	/*if len(markethistories.MarketHistories) > 0 {
		for i := 0; i < len(markethistories.MarketHistories); i++ {
			/*fmt.Println("AlbionId: " + markethistories.AlbionId)
			//fmt.Println("AlbionId Converted: " + string(state.marketHistoryIDLookup[markethistories.AlbionId]))
			fmt.Println("LocationId: " + string(markethistories.LocationId))
			fmt.Println("QualityLevel: " + markethistories.QualityLevel)
			fmt.Println("Timescale: " + markethistories.Timescale)
			fmt.Println("ItemAmount: " + markethistories.MarketHistories[i].ItemAmount)
			fmt.Println("SilverAmount: " + markethistories.MarketHistories[i].SilverAmount)
			fmt.Println("Timestamp: " + markethistories.MarketHistories[i].Timestamp)
			fmt.Println("\n")
		}

		var locationString string
		if markethistories.LocationId == 2004 || markethistories.LocationId == 2002 {
			locationString = "Bridgewatch"
		} else {
			locationString = "Unidentified"
		}
		folderPath := "marketdatastorage"
		itemTypeString := "idkItemType"
		filePath := folderPath + "/" + itemTypeString + "_" + locationString + "_" + "MarketHistory" + ".txt"
		if err := os.WriteFile(filePath, data, 0666); err != nil {
			log.Fatal(err)
		}

	}*/

	if len(orders.Orders) > 0 {
		typeOfItems := orders.Orders[0].ItemTypeId
		locationOfItems := orders.Orders[0].LocationId
		//Probably don't need to check quality
		//qualityOfItems := orders.Orders[0].QualityLevel
		enchantmentOfItems := orders.Orders[0].EnchantmentLevel
		auctionTypeOfItems := orders.Orders[0].AuctionType
		consistentData := "True"
		// Iterate through every order within our array
		// Then check each item is consistent
		for i := 0; i < len(orders.Orders); i++ {
			if typeOfItems != orders.Orders[i].ItemTypeId || locationOfItems != orders.Orders[i].LocationId || enchantmentOfItems != orders.Orders[i].EnchantmentLevel || auctionTypeOfItems != orders.Orders[i].AuctionType {
				consistentData = "False"
			}
			/*fmt.Println("Id: " + orders.Orders[i].Id)
			fmt.Println("ItemTypeId: " + orders.Orders[i].ItemTypeId)
			fmt.Println("ItemGroupTypeId: " + orders.Orders[i].ItemGroupTypeId)
			fmt.Println("LocationId: " + string(orders.Orders[i].LocationId))
			fmt.Println("QualityLevel: " + orders.Orders[i].QualityLevel)
			fmt.Println("EnchantmentLevel: " + orders.Orders[i].EnchantmentLevel)
			fmt.Println("UnitPriceSilver: " + orders.Orders[i].UnitPriceSilver)
			fmt.Println("Amount: " + orders.Orders[i].Amount)
			fmt.Println("AuctionType: " + orders.Orders[i].AuctionType)
			fmt.Println("Expires: " + orders.Orders[i].Expires)*/
		}

		//If all of the items match, save data
		if consistentData == "True" {
			//Change the portal town IDs to the main market IDs
			if orders.Orders[0].LocationId == 2301 {
				orders.Orders[0].LocationId = 2004 //Bridgewatch
			} /* else if orders.Orders[0].LocationId == ???? {
				orders.Orders[0].LocationId = 1002 //Lymhurst
			} else if orders.Orders[0].LocationId == ???? {
				orders.Orders[0].LocationId = 4002 //Fort Sterling
			} else if orders.Orders[0].LocationId == ???? {
				orders.Orders[0].LocationId = 3008 //Martlock
			} else if orders.Orders[0].LocationId == ???? {
				orders.Orders[0].LocationId = ???? //Thetford
			} else {
				orders.Orders[0].LocationId = "Unidentified"
			}*/

			//Get current time
			currentTime := time.Now()
			formattedTime := currentTime.Format(time.RFC3339Nano)
			log.Info("Current time: ", formattedTime)
			//Add record into db
			result, err := db.Exec("INSERT INTO traderecords (TradeId, ItemTypeId, ItemGroupTypeId, LocationId, QualityLevel, EnchantmentLevel, UnitPriceSilver, Amount, AuctionType, Expires, LastUpdated) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE TradeId=?, UnitPriceSilver=?, Amount=?, Expires=?, LastUpdated=?", orders.Orders[0].Id, orders.Orders[0].ItemTypeId, orders.Orders[0].ItemGroupTypeId, orders.Orders[0].LocationId, orders.Orders[0].QualityLevel, orders.Orders[0].EnchantmentLevel, orders.Orders[0].UnitPriceSilver, orders.Orders[0].Amount, orders.Orders[0].AuctionType, orders.Orders[0].Expires, formattedTime /*Splitting here*/, orders.Orders[0].Id, orders.Orders[0].UnitPriceSilver, orders.Orders[0].Amount, orders.Orders[0].Expires, formattedTime)
			if err != nil {
				fmt.Errorf("addRecord: %v", err)
			}
			_, err3 := result.LastInsertId()
			if err3 != nil {
				fmt.Errorf("addRecord: %v", err3)
			}
		} else {
			log.Error("Data does not contain consistent data, it was not written to file.")
		}
	}
	/*sendMsgToUploaders(data, topic, dis.publicUploaders)
	sendMsgToUploaders(data, topic, dis.privateUploaders)

	// If websockets are enabled, send the data there too
	if ConfigGlobal.EnableWebsockets {
		//sendMsgToWebSockets(data, topic)
	}*/
}

func sendMsgToPrivateUploaders(upload lib.PersonalizedUpload, topic string, state *albionState) {
	if ConfigGlobal.DisableUpload {
		log.Info("Upload is disabled.")
		return
	}

	// TODO: Re-enable this when issue #14 is fixed
	// Will personalize with blanks for now in order to allow people to see the format
	// if state.CharacterName == "" || state.CharacterId == "" {
	// 	log.Error("The player name or id has not been set. Please restart the game and make sure the client is running.")
	// 	notification.Push("The player name or id has not been set. Please restart the game and make sure the client is running.")
	// 	return
	// }

	upload.Personalize(state.CharacterId, state.CharacterName)

	data, err := json.Marshal(upload)
	if err != nil {
		log.Errorf("Error while marshalling payload for %v: %v", err, topic)
		return
	}

	if len(dis.privateUploaders) > 0 {
		sendMsgToUploaders(data, topic, dis.privateUploaders)
	}

	// If websockets are enabled, send the data there too
	if ConfigGlobal.EnableWebsockets {
		sendMsgToWebSockets(data, topic)
	}
}

func sendMsgToUploaders(msg []byte, topic string, uploaders []uploader) {
	if ConfigGlobal.DisableUpload {
		log.Info("Upload is disabled.")
		return
	}

	for _, u := range uploaders {
		u.sendToIngest(msg, topic)
	}
}

func runHTTPServer() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(wsHub, w, r)
	})

	err := http.ListenAndServe(":8099", nil)

	if err != nil {
		log.Panic("ListenAndServe: ", err)
	}
}

func sendMsgToWebSockets(msg []byte, topic string) {
	// TODO (gradius): send JSON data with topic string
	// TODO (gradius): this seems super hacky, and I'm sure there's a better way.
	var result string
	result = "{\"topic\": \"" + topic + "\", \"data\": " + string(msg) + "}"
	wsHub.broadcast <- []byte(result)
}
