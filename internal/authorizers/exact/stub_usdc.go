// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package exact

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ExactMetaData contains all meta data concerning the Exact contract.
var ExactMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"transferWithAuthorization\",\"outputs\":[],\"stateMutability\":\"pure\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f5ffd5b506102488061001c5f395ff3fe608060405234801561000f575f5ffd5b5060043610610029575f3560e01c8063e3ee160e1461002d575b5f5ffd5b6100476004803603810190610042919061014e565b610049565b005b505050505050505050565b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61008182610058565b9050919050565b61009181610077565b811461009b575f5ffd5b50565b5f813590506100ac81610088565b92915050565b5f819050919050565b6100c4816100b2565b81146100ce575f5ffd5b50565b5f813590506100df816100bb565b92915050565b5f819050919050565b6100f7816100e5565b8114610101575f5ffd5b50565b5f81359050610112816100ee565b92915050565b5f60ff82169050919050565b61012d81610118565b8114610137575f5ffd5b50565b5f8135905061014881610124565b92915050565b5f5f5f5f5f5f5f5f5f6101208a8c03121561016c5761016b610054565b5b5f6101798c828d0161009e565b995050602061018a8c828d0161009e565b985050604061019b8c828d016100d1565b97505060606101ac8c828d016100d1565b96505060806101bd8c828d016100d1565b95505060a06101ce8c828d01610104565b94505060c06101df8c828d0161013a565b93505060e06101f08c828d01610104565b9250506101006102028c828d01610104565b915050929598509295985092959856fea26469706673582212208999dc7f5efa1aebe80fb8738e5bb35a4cf95f30a2b3ef56ac837b483af9deeb64736f6c634300081e0033",
}

// ExactABI is the input ABI used to generate the binding from.
// Deprecated: Use ExactMetaData.ABI instead.
var ExactABI = ExactMetaData.ABI

// ExactBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ExactMetaData.Bin instead.
var ExactBin = ExactMetaData.Bin

// DeployExact deploys a new Ethereum contract, binding an instance of Exact to it.
func DeployExact(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Exact, error) {
	parsed, err := ExactMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ExactBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Exact{ExactCaller: ExactCaller{contract: contract}, ExactTransactor: ExactTransactor{contract: contract}, ExactFilterer: ExactFilterer{contract: contract}}, nil
}

// Exact is an auto generated Go binding around an Ethereum contract.
type Exact struct {
	ExactCaller     // Read-only binding to the contract
	ExactTransactor // Write-only binding to the contract
	ExactFilterer   // Log filterer for contract events
}

// ExactCaller is an auto generated read-only Go binding around an Ethereum contract.
type ExactCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExactTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ExactTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExactFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ExactFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExactSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ExactSession struct {
	Contract     *Exact            // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ExactCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ExactCallerSession struct {
	Contract *ExactCaller  // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// ExactTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ExactTransactorSession struct {
	Contract     *ExactTransactor  // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ExactRaw is an auto generated low-level Go binding around an Ethereum contract.
type ExactRaw struct {
	Contract *Exact // Generic contract binding to access the raw methods on
}

// ExactCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ExactCallerRaw struct {
	Contract *ExactCaller // Generic read-only contract binding to access the raw methods on
}

// ExactTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ExactTransactorRaw struct {
	Contract *ExactTransactor // Generic write-only contract binding to access the raw methods on
}

// NewExact creates a new instance of Exact, bound to a specific deployed contract.
func NewExact(address common.Address, backend bind.ContractBackend) (*Exact, error) {
	contract, err := bindExact(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Exact{ExactCaller: ExactCaller{contract: contract}, ExactTransactor: ExactTransactor{contract: contract}, ExactFilterer: ExactFilterer{contract: contract}}, nil
}

// NewExactCaller creates a new read-only instance of Exact, bound to a specific deployed contract.
func NewExactCaller(address common.Address, caller bind.ContractCaller) (*ExactCaller, error) {
	contract, err := bindExact(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ExactCaller{contract: contract}, nil
}

// NewExactTransactor creates a new write-only instance of Exact, bound to a specific deployed contract.
func NewExactTransactor(address common.Address, transactor bind.ContractTransactor) (*ExactTransactor, error) {
	contract, err := bindExact(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ExactTransactor{contract: contract}, nil
}

// NewExactFilterer creates a new log filterer instance of Exact, bound to a specific deployed contract.
func NewExactFilterer(address common.Address, filterer bind.ContractFilterer) (*ExactFilterer, error) {
	contract, err := bindExact(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ExactFilterer{contract: contract}, nil
}

// bindExact binds a generic wrapper to an already deployed contract.
func bindExact(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ExactMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Exact *ExactRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Exact.Contract.ExactCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Exact *ExactRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Exact.Contract.ExactTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Exact *ExactRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Exact.Contract.ExactTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Exact *ExactCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Exact.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Exact *ExactTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Exact.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Exact *ExactTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Exact.Contract.contract.Transact(opts, method, params...)
}

// TransferWithAuthorization is a free data retrieval call binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address , address , uint256 , uint256 , uint256 , bytes32 , uint8 , bytes32 , bytes32 ) pure returns()
func (_Exact *ExactCaller) TransferWithAuthorization(opts *bind.CallOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 *big.Int, arg5 [32]byte, arg6 uint8, arg7 [32]byte, arg8 [32]byte) error {
	var out []interface{}
	err := _Exact.contract.Call(opts, &out, "transferWithAuthorization", arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)

	if err != nil {
		return err
	}

	return err

}

// TransferWithAuthorization is a free data retrieval call binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address , address , uint256 , uint256 , uint256 , bytes32 , uint8 , bytes32 , bytes32 ) pure returns()
func (_Exact *ExactSession) TransferWithAuthorization(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 *big.Int, arg5 [32]byte, arg6 uint8, arg7 [32]byte, arg8 [32]byte) error {
	return _Exact.Contract.TransferWithAuthorization(&_Exact.CallOpts, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
}

// TransferWithAuthorization is a free data retrieval call binding the contract method 0xe3ee160e.
//
// Solidity: function transferWithAuthorization(address , address , uint256 , uint256 , uint256 , bytes32 , uint8 , bytes32 , bytes32 ) pure returns()
func (_Exact *ExactCallerSession) TransferWithAuthorization(arg0 common.Address, arg1 common.Address, arg2 *big.Int, arg3 *big.Int, arg4 *big.Int, arg5 [32]byte, arg6 uint8, arg7 [32]byte, arg8 [32]byte) error {
	return _Exact.Contract.TransferWithAuthorization(&_Exact.CallOpts, arg0, arg1, arg2, arg3, arg4, arg5, arg6, arg7, arg8)
}
