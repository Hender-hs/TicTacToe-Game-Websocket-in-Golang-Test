package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)



type AddMovePayload struct {
	UserMoveId string
	A1 string
	A2 string
	A3 string
	B1 string
	B2 string
	B3 string
	C1 string
	C2 string
	C3 string
}

type AddGuestPayload struct {
	Room_id string 
	Host string
	Host_id string
}

type AddGuestBodyMsg struct {
	RoomId string
	ReqType string
	Content AddGuestPayload
}

type AddMoveBodyMsg struct {
	RoomId string
	ReqType string
	Content map[string]interface{}
}




type RoomStateResponse struct {
	_id string
	Host map[string]interface{}
	Created_at string
	Updated_at string  
}



// var openedRooms = map[string]string{}

const DbApiUrl string = "http://127.0.0.1:5000/"

var wsupgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
	WriteBufferSize: 1024,
}

func cleanOriginHeader(ctx *gin.Context) {
	delete(ctx.Request.Header, "Origin")
}

func getClientReq(connection *websocket.Conn) (messageType int, p []byte, err error) {
	return connection.ReadMessage()
}


func parseByteToMap(content []byte) AddMoveBodyMsg {
	var output AddMoveBodyMsg
	
	err := json.Unmarshal(content, &output)

	// target := AddMoveBodyMsg{reqType: a["reqType"], content: a["content"], roomId: a["roomId"]}

	if err != nil {
		fmt.Println("func parseByteToMap; Error: ", err)
	}
	fmt.Println("parseByteToMap - content: ", string(content))
	fmt.Println("parseByteToMap - output: ", output)
	
	return output
}


func parseMapToByte(content AddMoveBodyMsg ) []byte {
	fmt.Println("content: ", content)
	parsedJSON, _ := json.Marshal(content)
	return parsedJSON
}

func parseResponseMapToByte(content AddMoveBodyMsg ) []byte {
	fmt.Println("content: ", content)
	parsedJSON, _ := json.Marshal(content)
	return parsedJSON
}


func getRoomState(roomId string) AddMoveBodyMsg {

	res, requestError := http.Get(DbApiUrl + "rooms/" + roomId)

	if requestError != nil {
		fmt.Println("error when making the room request")	
	}
	defer res.Body.Close()

	byteBody, ioReadingError := io.ReadAll(res.Body) 

	if ioReadingError != nil {
		fmt.Println("error reading the request response")
	}

	body := parseByteToMap(byteBody)

	return body
}


func addUserMove(roomId string, content AddMoveBodyMsg) AddMoveBodyMsg {
	
	// This varible is to create a type of "io.Reader"
	// so that the "http.Post" method can work.
	reqContent := string(parseMapToByte(content))

	res, requestError := http.Post(DbApiUrl + "add_move" + roomId, "json", strings.NewReader(reqContent))

	if requestError != nil {
		fmt.Println("error when adding the user move")	
	}

	byteBody, ioReadingError := io.ReadAll(res.Body) 

	if ioReadingError != nil {
		fmt.Println("error reading the request response")
	}

	body := parseByteToMap(byteBody)

	return body
}


func addGuest(roomId string, content AddMoveBodyMsg) AddMoveBodyMsg {

	// This varible is to create a type of "io.Reader"
	// so that the "http.Post" method can work.


	reqContent := string(parseMapToByte(content))

	res, requestError := http.Post(DbApiUrl + "add_guest" + roomId, "json", strings.NewReader(reqContent))

	if requestError != nil {
		fmt.Println("error when adding the user move")	
	}

	byteBody, ioReadingError := io.ReadAll(res.Body) 

	if ioReadingError != nil {
		fmt.Println("error reading the request response")
	}

	body := parseByteToMap(byteBody)

	return body
}



func communicateWithApi(clientReq AddMoveBodyMsg) AddMoveBodyMsg {
	var resOutput AddMoveBodyMsg

	switch clientReq.ReqType {
	case "add_move":
		resOutput = addUserMove(clientReq.RoomId, clientReq)
	case "add_guest":
		resOutput = addGuest(clientReq.RoomId, clientReq)
	case "get_room_state":
		resOutput = getRoomState(clientReq.RoomId)
	}

	return resOutput
}


func WSbodyValidation(body AddMoveBodyMsg) (err string) {
	if body.RoomId == "" {
		return "Missing roomId field"
	}
	if body.ReqType == "" {
		return "Missing reqType field"
	}

	// if body.Content == "" {
	// 	// TODO: make a deep validation of this field ("content" key).
	// 	return "Missing content field"
	// }



	return ""
}

func typeBodyMsg(body *map[string]string) {
}


func sendMsgToClient(connection *websocket.Conn, msgType int, response []byte) error {
	err := connection.WriteMessage(msgType, response);
	return err
}


func wshandler(ctx *gin.Context) {

	cleanOriginHeader(ctx)

	connection, connErr := wsupgrader.Upgrade(ctx.Writer, ctx.Request, nil)

	if connErr != nil {
		fmt.Printf("Failed to set websocket upgrade: %+v \n", connErr)
		return
	}

	for {
		msgType, msg, _ := getClientReq(connection)

		jsonMap := parseByteToMap(msg)
		fmt.Println("jsonMap: ", jsonMap)
		err := WSbodyValidation(jsonMap)

		var response []byte

		if err != "" {
			response = []byte(err)
		}

		apiResponse := communicateWithApi(jsonMap)

		fmt.Println("apiResponse: ", apiResponse)

		response = parseMapToByte(apiResponse)

		sendMsgError := sendMsgToClient(connection, msgType, response);
		if sendMsgError != nil {
			fmt.Println("sendMsgError: ", sendMsgError)
		}
	}
}

func routes(engine *gin.Engine) {
	engine.GET("/ws", wshandler)
}

func initServer(engine *gin.Engine, port string) {
	engine.Run(port)
}

func main() {
	engine := gin.Default()
	engine.SetTrustedProxies([]string{ "127.0.0.1" })
	routes(engine)

	serverPort := "0.0.0.0:3000"
	initServer(engine, serverPort)
	return
}