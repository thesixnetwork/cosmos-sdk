#!/bin/bash
set -e # Exit on error

# =====================================================
# CONFIGURATION SECTION - Easy to modify parameters
# =====================================================

# Chain configuration
CHAINID="testnet"
MONIKER="${1:-mynode}"
KEYRING="test"
KEYALGO="secp256k1"
STAKE_HOME=~/.simapp
LOGLEVEL="info"
VAL_MODE=$2
TRACE="" # Set to "--trace" for tracing

if [ -z "$VAL_MODE" ]; then
  VAL_MODE=0
fi


# Token denominations
STAKING_TOKEN="ustake"

# =====================================================
# KEY ADDRESS MAPPING - Important for matching config.yml
# =====================================================

# These are the key addresses from the working genesis
SUPER_ADMIN_ADDRESS="cosmos1t3p2vzd7w036ahxf4kefsc9sn24pvlqpcktgg7"

# =====================================================
# MNEMONICS SECTION - From config.yml only
# =====================================================

# Mnemonics from config.yml
ALICE_MNEMONIC="history perfect across group seek acoustic delay captain sauce audit carpet tattoo exhaust green there giant cluster want pond bulk close screen scissors remind"
BOB_MNEMONIC="limb sister humor wisdom elephant weasel beyond must any desert glance stem reform soccer include chest chef clerk call popular display nerve priority venture"
SUPER_ADMIN_MNEMONIC="expect peace defense conduct virtual flight flip unit equip solve broccoli protect shed group else useless tree such tornado minimum decade tower warfare galaxy"

# =====================================================
# VALIDATION SECTION
# =====================================================
echo "Starting initialization of $CHAINID testnet with validator bob..."

# Validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# =====================================================
# SETUP SECTION
# =====================================================
echo "Setting up environment..."

# Reinstall daemon
rm -rf ${STAKE_HOME}
rm go.sum && touch go.sum
go mod tidy
make install

# Set client config
simd config set client chain-id $CHAINID --home ${STAKE_HOME}
simd config set client keyring-backend $KEYRING --home ${STAKE_HOME}

# =====================================================
# KEY MANAGEMENT SECTION
# =====================================================
echo "Importing keys from config.yml..."

# Import keys
echo $ALICE_MNEMONIC | simd keys add alice --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING} --algo ${KEYALGO}
echo $BOB_MNEMONIC | simd keys add bob --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING} --algo ${KEYALGO}
echo $SUPER_ADMIN_MNEMONIC | simd keys add super-admin --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING} --algo ${KEYALGO}

# =====================================================
# CHAIN INITIALIZATION SECTION
# =====================================================
echo "Initializing chain with moniker: $MONIKER and chain-id: $CHAINID"
simd init $MONIKER --chain-id $CHAINID --home ${STAKE_HOME}

# =====================================================
# GENESIS CONFIGURATION SECTION
# =====================================================
echo "Configuring genesis..."

# Function to update genesis using jq
update_genesis() {
    cat ${STAKE_HOME}/config/genesis.json | jq "$1" > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
}

# Change parameter token denominations from stake to ustake
update_genesis '.app_state["staking"]["params"]["bond_denom"]="'$STAKING_TOKEN'"'
update_genesis '.app_state["crisis"]["constant_fee"]["denom"]="'$STAKING_TOKEN'"'
update_genesis '.app_state["crisis"]["constant_fee"]["amount"]="1000"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="'$STAKING_TOKEN'"'
update_genesis '.app_state["gov"]["deposit_params"]["min_deposit"][0]["amount"]="1000000"'
update_genesis '.app_state["inflation"]["params"]["mint_denom"]="'$STAKING_TOKEN'"'
update_genesis '.app_state["mint"]["params"]["mint_denom"]="'$STAKING_TOKEN'"'

# Bank configuration
update_genesis '.app_state.bank.params.default_send_enabled = true'

# Token metadata - identical to working genesis
update_genesis '.app_state.bank.denom_metadata[0] = {
  "description": "The native staking token of the stake Protocol.",
  "denom_units": [
    {"denom": "ustake","exponent": 0,"aliases": ["microstake"]},
    {"denom": "mstake","exponent": 3,"aliases": ["millistake"]},
    {"denom": "stake","exponent": 6,"aliases": []}
  ],
  "base": "ustake",
  "display": "stake",
  "name": "stake token",
  "symbol": "stake"
}'

# Validator approval configuration
update_genesis '.app_state.staking.validator_approval = {
  "approver_address": "'$SUPER_ADMIN_ADDRESS'",
  "enabled": false
}'

update_genesis '.app_state.staking.params.max_validators = 3'
update_genesis '.app_state.staking.params.unbonding_time = "300s"'

# Governance configuration - Match working genesis
update_genesis '.app_state.gov.deposit_params.max_deposit_period = "300s"'
update_genesis '.app_state.gov.voting_params.voting_period = "300s"'


# =====================================================
# CONFIG.TOML CONFIGURATION - Match ignite settings
# =====================================================
echo "Configuring config.toml..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    # RPC settings
    sed -i '' 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["*",\]/g' ${STAKE_HOME}/config/config.toml
    # Consensus settings
    sed -i '' 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "500ms"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "1s"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "500ms"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "1s"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "500ms"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "10s"/g' ${STAKE_HOME}/config/config.toml
    sed -i '' 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' ${STAKE_HOME}/config/config.toml
else
    # RPC settings
    sed -i 's/laddr = "tcp:\/\/127.0.0.1:26657"/laddr = "tcp:\/\/0.0.0.0:26657"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/cors_allowed_origins = \[\]/cors_allowed_origins = \["*",\]/g' ${STAKE_HOME}/config/config.toml
    # Consensus settings
    sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_propose = "3s"/timeout_propose = "30s"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_propose_delta = "500ms"/timeout_propose_delta = "5s"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_prevote = "1s"/timeout_prevote = "1s"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "500ms"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_precommit = "1s"/timeout_precommit = "1s"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "500ms"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "10s"/g' ${STAKE_HOME}/config/config.toml
    sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' ${STAKE_HOME}/config/config.toml
fi

# =====================================================
# APP.TOML CONFIGURATION - Match ignite settings
# =====================================================
echo "Configuring app.toml to match ignite settings..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    # API configuration
    sed -i '' 's/enable = false/enable = true/g' ${STAKE_HOME}/config/app.toml
    sed -i '' 's/swagger = false/swagger = true/g' ${STAKE_HOME}/config/app.toml
    sed -i '' 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' ${STAKE_HOME}/config/app.toml
    sed -i '' 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:1317"/g' ${STAKE_HOME}/config/app.toml

    # gRPC configuration
    sed -i '' 's/enable = false/enable = true/g' ${STAKE_HOME}/config/app.toml
    sed -i '' 's/address = "localhost:9090"/address = "0.0.0.0:9090"/g' ${STAKE_HOME}/config/app.toml
    sed -i '' 's/address = "0.0.0.0:9091"/address = "0.0.0.0:9091"/g' ${STAKE_HOME}/config/app.toml
    sed -i '' 's/enable = false/enable = true/g' ${STAKE_HOME}/config/app.toml
else
    # API configuration
    sed -i 's/enable = false/enable = true/g' ${STAKE_HOME}/config/app.toml
    sed -i 's/swagger = false/swagger = true/g' ${STAKE_HOME}/config/app.toml
    sed -i 's/enabled-unsafe-cors = false/enabled-unsafe-cors = true/g' ${STAKE_HOME}/config/app.toml
    sed -i 's/address = "tcp:\/\/localhost:1317"/address = "tcp:\/\/0.0.0.0:1317"/g' ${STAKE_HOME}/config/app.toml

    # gRPC configuration
    sed -i 's/enable = false/enable = true/g' ${STAKE_HOME}/config/app.toml
    sed -i 's/address = "localhost:9090"/address = "0.0.0.0:9090"/g' ${STAKE_HOME}/config/app.toml
    sed -i 's/address = "0.0.0.0:9091"/address = "0.0.0.0:9091"/g' ${STAKE_HOME}/config/app.toml
    sed -i 's/enable = false/enable = true/g' ${STAKE_HOME}/config/app.toml
fi

# =====================================================
# ACCOUNT ALLOCATION SECTION
# =====================================================
echo "Allocating genesis accounts..."

# Function to add genesis accounts
add_genesis_account() {
    local address=$1
    local amount=$2
    simd genesis add-genesis-account $address $amount --home ${STAKE_HOME}
}
add_genesis_account $(simd keys show -a alice --keyring-backend ${KEYRING} --home ${STAKE_HOME}) "10000000000000${STAKING_TOKEN}"
add_genesis_account $(simd keys show -a bob --keyring-backend ${KEYRING} --home ${STAKE_HOME})  "11000000000000${STAKING_TOKEN}"
add_genesis_account $(simd keys show -a super-admin --keyring-backend ${KEYRING} --home ${STAKE_HOME}) "1000000000000${STAKING_TOKEN}"
# =====================================================
# GENTX SECTION
# =====================================================
echo "Creating and collecting gentxs with bob as validator..."

if [ "$VAL_MODE" = "0" ]; then
  simd genesis gentx bob 1000000000000ustake --min-self-delegation="10000000000" --validator-mode=0 --min-delegation="10000000000" --keyring-backend $KEYRING --chain-id $CHAINID
elif [ "$VAL_MODE" = "1" ]; then
  simd genesis gentx bob 20000000000ustake --min-self-delegation="10000000000" --validator-mode=1 --min-delegation="10000000000" --delegation-increment="10000000000" --max-license=10 --keyring-backend $KEYRING --chain-id $CHAINID --home ${STAKE_HOME}
elif [ "$VAL_MODE" = "2" ]; then
  simd genesis gentx bob 1000000000000ustake --min-self-delegation="10000000000" --validator-mode=2 --min-delegation="10000000000" --keyring-backend $KEYRING --chain-id $CHAINID
else
  echo "Invalid validator mode: $VAL_MODE. Using default mode 0."
  simd genesis gentx bob 1000000000000ustake --min-self-delegation="10000000000" --validator-mode=0 --min-delegation="10000000000" --keyring-backend $KEYRING --chain-id $CHAINID
fi

# Collect genesis tx
simd genesis collect-gentxs --home ${STAKE_HOME}

# Run this to ensure everything worked and that the genesis file is setup correctly
simd genesis validate-genesis --home ${STAKE_HOME}

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
simd start --rpc.laddr "tcp://0.0.0.0:26657" --api.enable true $TRACE --log_level ${LOGLEVEL} --home ${STAKE_HOME}
