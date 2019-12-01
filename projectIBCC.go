package projectIBCC

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"

	chain "github.com/aneesulanwar/projectIBC"
)

type Node struct {
	//Node represent the person
	Name    string
	Address string
	Port    string
}

type CAddress struct {
	//CAddress to store addresses of connected nodes
	Name    string
	Address string
	Port    string
}
type NetworkTrans struct {
	Name        string
	Data        string
	Block       *chain.Block
	Bchain      *chain.Block
	Addresses   []CAddress
	Transaction chain.Transaction
}

var Nodes []CAddress
var Leader CAddress
var Fupdate bool

func HandleConnection(con net.Conn, thisNode CAddress, chainHead **chain.Block, leader CAddress) {

	var recvdBlock NetworkTrans
	dec := gob.NewDecoder(con)
	err := dec.Decode(&recvdBlock)
	if err != nil {
		// handle error
	}
	if recvdBlock.Name == "FirstUpdate" {
		*(chainHead) = recvdBlock.Bchain
		Nodes = recvdBlock.Addresses
		Nodes = append(Nodes, leader)
		chain.ListBlocks(*(chainHead))
		fmt.Println(Nodes)
		Fupdate = true
	}

	if recvdBlock.Name == "Validate" {
		Validate(recvdBlock.Transaction, thisNode, chainHead)
	}

	if recvdBlock.Name == "ValidateBlock" {
		ValidateBlock(recvdBlock.Block, chainHead)
	}
}

func Validate(transaction chain.Transaction, thisNode CAddress, chainHead **chain.Block) {

	fmt.Println("received Validate Transaction")

	validTransaction := true
	var temp *chain.Block
	temp = *(chainHead)
	amount := 0.0
	for temp.PrevPointer != nil {
		i := 0
		for i < len(temp.Transactions) {
			if temp.Transactions[i].To == transaction.From {
				amount += temp.Transactions[i].Bcoins
			}

			if temp.Transactions[i].From == transaction.From {
				amount -= temp.Transactions[i].Bcoins
			}
			i = i + 1
		}
		temp = temp.PrevPointer
	}
	i := 0
	for i < len(temp.Transactions) {
		if temp.Transactions[i].To == transaction.From {
			amount += temp.Transactions[i].Bcoins
		}
		if temp.Transactions[i].From == transaction.From {
			amount -= temp.Transactions[i].Bcoins
		}
		i = i + 1
	}

	if amount < transaction.Bcoins {
		/*fmt.Println("Invalid Transaction")

		fmt.Println("Do you want to add it, yes/no")

		var valid string
		fmt.Scan(&valid)
		for valid != "yes" && valid != "no" {
			fmt.Println("Do you want to add it, yes/no")
			fmt.Scan(&valid)
		}
		fmt.Println("input is : ", valid)
		if valid == "yes" {
			fmt.Println("enter into yes block")
			validTransaction = true
		} else if valid == "no" {
			fmt.Println("enter into No block")
			validTransaction = false
		}
		*/

		fmt.Println("Invalid Transaction")
		decision := rand.Intn(4)
		if decision == 0 {
			validTransaction = true
		} else {
			validTransaction = true
		}

		fmt.Println("Decision is ", validTransaction)

	}

	if validTransaction {
		fmt.Println("Valid Transaction")
		var newTran chain.Transaction
		newTran.To = thisNode.Name
		newTran.From = "mining"
		newTran.Bcoins = 100
		var Block chain.Block
		Block.Transactions = append(Block.Transactions, newTran)
		Block.Transactions = append(Block.Transactions, transaction)
		Block.DeriveHash()
		toAdd := &Block
		Block1 := &Block
		/////
		temp := *(chainHead)
		toAdd.PrevBlockHash = temp.Hash
		toAdd.PrevPointer = temp
		ValidateBlock(toAdd, chainHead)
		Block1.PrevPointer = temp
		Block1.PrevBlockHash = temp.Hash
		//*(chainHead) = toAdd
		/////
		for i := 0; i < len(Nodes); i++ {
			Propagate(Block1, Nodes[i])
		}

		chain.ListBlocks(*(chainHead))
	} else {
		fmt.Println("InValid Transaction")
		var newTran chain.Transaction
		newTran.To = thisNode.Name
		newTran.From = "mining"
		newTran.Bcoins = 100
		var Block chain.Block
		Block.Transactions = append(Block.Transactions, newTran)
		Block.DeriveHash()
		toAdd := &Block
		Block1 := &Block
		/////
		temp := *(chainHead)
		toAdd.PrevBlockHash = temp.Hash
		toAdd.PrevPointer = temp
		ValidateBlock(toAdd, chainHead)
		Block1.PrevPointer = temp
		Block1.PrevBlockHash = temp.Hash
		//*(chainHead) = toAdd
		/////
		for i := 0; i < len(Nodes); i++ {
			Propagate(Block1, Nodes[i])
		}

		chain.ListBlocks(*(chainHead))
	}

}

func Propagate(block *chain.Block, node CAddress) {
	conn, err := net.Dial("tcp", node.Address+":"+node.Port)
	if err != nil {
		// handle error
		log.Println(err)
		fmt.Println("error in connection")

	}

	var blck NetworkTrans
	blck.Name = "ValidateBlock"
	blck.Block = block
	gobEncoder := gob.NewEncoder(conn)
	err1 := gobEncoder.Encode(blck)
	if err1 != nil {
		log.Println(err)
	}
}

func ValidateBlock(block *chain.Block, chainHead **chain.Block) {
	//fmt.Println("received Validate Block")
	validb := true //if block is valid

	var tempv *chain.Block
	tempv = *(chainHead)
	tempb := block
	amount := 0.0

	for t := 0; t < len(tempb.Transactions); t++ {

		if tempb.Transactions[t].From != "mining" {
			for tempv.PrevPointer != nil {
				i := 0
				for i < len(tempv.Transactions) {
					if tempv.Transactions[i].To == tempb.Transactions[t].From {
						amount += tempv.Transactions[i].Bcoins
					}
					if tempv.Transactions[i].From == tempb.Transactions[t].From {
						amount -= tempv.Transactions[i].Bcoins
					}
					i = i + 1
				}
				tempv = tempv.PrevPointer
			}
			i := 0
			for i < len(tempv.Transactions) {
				if tempv.Transactions[i].To == tempb.Transactions[t].From {
					amount += tempv.Transactions[i].Bcoins
				}
				if tempv.Transactions[i].From == tempb.Transactions[t].From {
					amount -= tempv.Transactions[i].Bcoins
				}
				i = i + 1
			}

			if amount < tempb.Transactions[t].Bcoins {
				validb = false
			}
		} else {
			if tempb.Transactions[t].Bcoins != 100 {
				validb = false
			}
		}
	}

	if validb {
		temp1 := block
		temp := *(chainHead)
		result := bytes.Compare(block.Hash, temp.Hash)
		result1 := bytes.Compare(block.PrevBlockHash, temp.PrevBlockHash)
		len1 := length(block)
		len2 := length(*(chainHead))
		if result != 0 || result1 != 0 || len1 > len2 {
			block.PrevBlockHash = temp.Hash
			block.PrevPointer = temp
			*(chainHead) = block
			i := 0
			for i < len(Nodes) {
				Propagate(temp1, Nodes[i])
				i = i + 1
			}

			chain.ListBlocks(*(chainHead))
		}
	}
}

func WantTransaction(beginer CAddress) {
	for {

		if Fupdate == true {
			fmt.Println("do you want to perform transaction?")
			var trans string
			fmt.Scan(&trans)
			if trans == "yes" {
				var wg sync.WaitGroup
				wg.Add(1)
				StartTransaction(beginer, &wg)
				wg.Wait()
			}
		}
	}
}

func StartTransaction(beginer CAddress, wg *sync.WaitGroup) {
	fmt.Println("enter the name of receiver")
	var receiver string
	fmt.Scan(&receiver)

	fmt.Println("enter the amount of Bcoins you want to transfer")
	var amount float64
	fmt.Scan(&amount)

	var newTrans chain.Transaction
	newTrans.To = receiver
	newTrans.From = beginer.Name
	newTrans.Bcoins = amount

	var newBlock NetworkTrans
	if newTrans.To == "stake" {
		newBlock.Name = "Stake"
		for amount > 100 {
			fmt.Println("enter the amount of stake again, can't have stake greater than 100")
			fmt.Scan(&amount)
		}
		newTrans.Bcoins = amount
	} else if newTrans.To == beginer.Name && newTrans.From == "stake" {
		newBlock.Name = "Reverse Stake"
	} else {
		newBlock.Name = "Validate"
	}
	newBlock.Transaction = newTrans

	conn, err := net.Dial("tcp", Leader.Address+":"+Leader.Port)
	if err != nil {
		// handle error
		log.Println(err)
		fmt.Println("error in connection")

	}
	gobEncoder := gob.NewEncoder(conn)
	err1 := gobEncoder.Encode(newBlock)
	if err1 != nil {
		log.Println(err)
	}

	defer wg.Done()
}

func length(block *chain.Block) int {
	temp := block
	len := 0
	for temp != nil {
		len++
		temp = temp.PrevPointer
	}

	return len
}
