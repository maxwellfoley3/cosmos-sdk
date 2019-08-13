package genutil

import (
	"encoding/json"
	"math/rand"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	sim "github.com/cosmos/cosmos-sdk/x/simulation"
)

var (
	_ module.AppModuleGenesis    = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModuleSimulation{}
)

// AppModuleBasic defines the basic application module used by the genutil module.
type AppModuleBasic struct{}

// Name returns the genutil module's name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterCodec registers the genutil module's types for the given codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {}

// DefaultGenesis returns default genesis state as raw bytes for the genutil
// module.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(GenesisState{})
}

// ValidateGenesis performs genesis state validation for the genutil module.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}
	return ValidateGenesis(data)
}

// RegisterRESTRoutes registers the REST routes for the genutil module.
func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// GetTxCmd returns no root tx command for the genutil module.
func (AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command { return nil }

// GetQueryCmd returns no root query command for the genutil module.
func (AppModuleBasic) GetQueryCmd(_ *codec.Codec) *cobra.Command { return nil }

//____________________________________________________________________________

// AppModuleSimulation defines the module simulation functions used by the genutil module.
type AppModuleSimulation struct{}

// RegisterStoreDecoder performs a no-op.
func (AppModuleSimulation) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// GenerateGenesisState creates a randomized GenState of the genutil module.
func (AppModuleSimulation) GenerateGenesisState(_ *codec.Codec, _ *rand.Rand, _ map[string]json.RawMessage) {}

// RandomizedParams doesn't create randomized genaccounts param changes for the simulator.
func (AppModuleSimulation) RandomizedParams(_ *codec.Codec, _ *rand.Rand) []sim.ParamChange {
	return nil
}

//____________________________________________________________________________

// AppModule implements an application module for the genutil module.
type AppModule struct {
	AppModuleBasic

	accountKeeper types.AccountKeeper
	stakingKeeper types.StakingKeeper
	deliverTx     deliverTxfn
}

// NewAppModule creates a new AppModule object
func NewAppModule(accountKeeper types.AccountKeeper,
	stakingKeeper types.StakingKeeper, deliverTx deliverTxfn) module.AppModule {

	return module.NewGenesisOnlyAppModule(AppModule{
		AppModuleBasic:      AppModuleBasic{},
		accountKeeper:       accountKeeper,
		stakingKeeper:       stakingKeeper,
		deliverTx:           deliverTx,
	})
}

// InitGenesis performs genesis initialization for the genutil module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	return InitGenesis(ctx, ModuleCdc, am.stakingKeeper, am.deliverTx, genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the genutil
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return nil
}
