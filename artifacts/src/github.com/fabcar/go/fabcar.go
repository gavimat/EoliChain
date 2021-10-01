package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// Aerogerador:  Define the Aerogerador structure, with 4 properties.  Structure tags are used by encoding/json library
type Aerogerador struct {
	Localizacao string `json:"localizacao"`
	Operador    string `json:"operador"`
	Status      string `json:"status"`
	Balanceado  string `json:"balanceado"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabcar_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryAerogerador":
		return s.queryAerogerador(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "criarAerogerador":
		return s.criarAerogerador(APIstub, args)
	case "queryAll":
		return s.queryAll(APIstub)
	case "alterarOperadorAerogerador":
		return s.alterarOperadorAerogerador(APIstub, args)

	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (s *SmartContract) queryAerogerador(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Número incorreto de argumentos. Esperado 1 argumento.")
	}

	aeroAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(aeroAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	aerogeradores := []Aerogerador{
		Aerogerador{Localizacao: "RS", Operador: "José", Status: "Em operacao", Balanceado: "Balanceado"},
		Aerogerador{Localizacao: "RS", Operador: "Matheus", Status: "Falha", Balanceado: "Sem dados"},
		Aerogerador{Localizacao: "RJ", Operador: "Eduardo", Status: "Em operacao", Balanceado: "Desbalanceado"},
		Aerogerador{Localizacao: "MG", Operador: "Ricardo", Status: "Em operacao", Balanceado: "Balanceado"},
	}

	i := 0
	for i < len(aerogeradores) {
		aeroAsBytes, _ := json.Marshal(aerogeradores[i])
		APIstub.PutState("AEROGERADOR"+strconv.Itoa(i), aeroAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) criarAerogerador(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Número incorreto de argumentos. Esperado 5 argumentos.")
	}

	var aerogerador = Aerogerador{Localizacao: args[1], Operador: args[2], Status: args[3], Balanceado: args[4]}

	aeroAsBytes, _ := json.Marshal(aerogerador)
	APIstub.PutState(args[0], aeroAsBytes)

	indexName := "owner~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{aerogerador.Operador, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(aeroAsBytes)
}

func (s *SmartContract) queryAll(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "AEROGERADOR0"
	endKey := "AEROGERADOR999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAll:\n %s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) alterarOperadorAerogerador(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Número incorreto de argumentos. Esperado 2 argumentos.")
	}

	aeroAsBytes, _ := APIstub.GetState(args[0])
	aero := Aerogerador{}

	json.Unmarshal(aeroAsBytes, &aero)
	aero.Operador = args[1]

	aeroAsBytes, _ = json.Marshal(aero)
	APIstub.PutState(args[0], aeroAsBytes)

	return shim.Success(aeroAsBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
