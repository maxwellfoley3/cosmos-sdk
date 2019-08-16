package simapp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	authsim "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	banksim "github.com/cosmos/cosmos-sdk/x/bank/simulation"
	distrsim "github.com/cosmos/cosmos-sdk/x/distribution/simulation"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	genaccsim "github.com/cosmos/cosmos-sdk/x/genaccounts/simulation"
	govsim "github.com/cosmos/cosmos-sdk/x/gov/simulation"
	mintsim "github.com/cosmos/cosmos-sdk/x/mint/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	slashingsim "github.com/cosmos/cosmos-sdk/x/slashing/simulation"
	stakingsim "github.com/cosmos/cosmos-sdk/x/staking/simulation"
	supplysim "github.com/cosmos/cosmos-sdk/x/supply/simulation"
)

// AppStateFn returns the initial application state using a genesis or the simulation parameters.
// It panics if the user provides files for both of them.
// If a file is not given for the genesis or the sim params, it creates a randomized one.
func AppStateFn(cdc *codec.Codec) simulation.AppStateFn {
	return func(r *rand.Rand, accs []simulation.Account, config simulation.Config,
	) (appState json.RawMessage, simAccs []simulation.Account, chainID string, genesisTimestamp time.Time) {

		if flagGenesisTimeValue == 0 {
			genesisTimestamp = simulation.RandTimestamp(r)
		} else {
			genesisTimestamp = time.Unix(flagGenesisTimeValue, 0)
		}

		switch {
		case config.ParamsFile != "" && config.GenesisFile != "":
			panic("cannot provide both a genesis file and a params file")

		case config.GenesisFile != "":
			appState, simAccs, chainID = AppStateFromGenesisFileFn(r, cdc, config)

		case config.ParamsFile != "":
			appParams := make(simulation.AppParams)
			bz, err := ioutil.ReadFile(config.ParamsFile)
			if err != nil {
				panic(err)
			}

			cdc.MustUnmarshalJSON(bz, &appParams)
			appState, simAccs, chainID = AppStateRandomizedFn(r, cdc, accs, genesisTimestamp, appParams)

		default:
			appParams := make(simulation.AppParams)
			appState, simAccs, chainID = AppStateRandomizedFn(r, cdc, accs, genesisTimestamp, appParams)
		}

		return appState, simAccs, chainID, genesisTimestamp
	}
}

// AppStateRandomizedFn creates calls each module's GenesisState generator function
// and creates
func AppStateRandomizedFn(
	r *rand.Rand, cdc *codec.Codec, accs []simulation.Account, genesisTimestamp time.Time, appParams simulation.AppParams,
) (json.RawMessage, []simulation.Account, string) {

	genesisState := NewDefaultGenesisState()
	numInitiallyBonded, amount := RandomizedSimulationParams(r, int64(len(accs)))

	genaccsim.RandomizedGenState(cdc, r, genesisState, accs, amount, numInitiallyBonded, genesisTimestamp)
	authsim.RandomizedGenState(cdc, r, genesisState)
	banksim.RandomizedGenState(cdc, r, genesisState)
	supplysim.RandomizedGenState(cdc, r, genesisState, accs, amount, numInitiallyBonded)
	govsim.RandomizedGenState(cdc, r, genesisState)
	mintsim.RandomizedGenState(cdc, r, genesisState)
	distrsim.RandomizedGenState(cdc, r, genesisState)
	stakingGen := stakingsim.RandomizedGenState(cdc, r, genesisState, accs, amount, numInitiallyBonded)
	slashingsim.RandomizedGenState(cdc, r, genesisState, stakingGen.Params.UnbondingTime)

	appState, err := cdc.MarshalJSON(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs, "simulation"
}

// AppStateFromGenesisFileFn util function to generate the genesis AppState
// from a genesis.json file
func AppStateFromGenesisFileFn(r *rand.Rand, cdc *codec.Codec, config simulation.Config) (
	genState json.RawMessage, newAccs []simulation.Account, chainID string) {

	bytes, err := ioutil.ReadFile(config.GenesisFile)
	if err != nil {
		panic(err)
	}

	var genesis tmtypes.GenesisDoc
	cdc.MustUnmarshalJSON(bytes, &genesis)

	var appState GenesisState
	cdc.MustUnmarshalJSON(genesis.AppState, &appState)

	accounts := genaccounts.GetGenesisStateFromAppState(cdc, appState)

	for _, acc := range accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		if _, err := r.Read(privkeySeed); err != nil {
			panic(err)
		}

		privKey := secp256k1.GenPrivKeySecp256k1(privkeySeed)

		// create simulator accounts
		simAcc := simulation.Account{PrivKey: privKey, PubKey: privKey.PubKey(), Address: acc.Address}
		newAccs = append(newAccs, simAcc)
	}

	return genesis.AppState, newAccs, genesis.ChainID
}

// RandomizedSimulationParams generate a random amount of initial stake coins
// and a random initially bonded number of accounts
func RandomizedSimulationParams(r *rand.Rand, numAccs int64) (numInitiallyBonded, initialStake int64) {
	initialStake = int64(r.Intn(1e12))
	numInitiallyBonded = int64(r.Intn(250))

	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	fmt.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  stake_per_account: "%d",
  initially_bonded_validators: "%d"
}
`, initialStake, numInitiallyBonded,
	)

	return
}
