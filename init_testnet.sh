MONIKER=$1
# if empty, default to "testnet"
if [ -z "$MONIKER" ]; then
  MONIKER="mynode"
fi
VALKEY=val1 # should be: export as docker env var
STAKE_HOME=~/.simapp
ALICE_MNEMONIC="history perfect across group seek acoustic delay captain sauce audit carpet tattoo exhaust green there giant cluster want pond bulk close screen scissors remind"
BOB_MNEMONIC="limb sister humor wisdom elephant weasel beyond must any desert glance stem reform soccer include chest chef clerk call popular display nerve priority venture"
VAL1_MNEMONIC="note base stone list envelope tail start forget alarm acoustic cook occur divert giant bike curtain chase shuffle fade glow capital slot file provide"
SUPER_ADMIN_MNEMONIC="expect peace defense conduct virtual flight flip unit equip solve broccoli protect shed group else useless tree such tornado minimum decade tower warfare galaxy"
KEY="mykey"
CHAINID="testnet"
KEYRING="test"
KEYALGO="secp256k1"
LOGLEVEL="info"
# to trace evm
#TRACE="--trace"
TRACE=""

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

# Reinstall daemon
rm -rf ~/.simapp*
make install LEDGER_ENABLED=true

# Set client config
simd config keyring-backend $KEYRING
simd config chain-id $CHAINID

# if $KEY exists it should be deleted
# mint to validator
echo $SUPER_ADMIN_MNEMONIC | simd keys add super-admin --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING} --algo ${KEYALGO}
echo $ALICE_MNEMONIC | simd keys add alice --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING} --algo ${KEYALGO}
echo $BOB_MNEMONIC | simd keys add bob --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING} --algo ${KEYALGO}
echo $VAL1_MNEMONIC | simd keys add val1 --recover --home ${STAKE_HOME} --keyring-backend ${KEYRING} --algo ${KEYALGO}

# Set moniker and chain-id for stake (Moniker can be anything, chain-id must be an integer)
simd init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to ustake
## from stake to ustake
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="ustake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="ustake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="ustake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["evm"]["params"]["evm_denom"]="astake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="ustake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="ustake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.bank.params.send_enabled[0] = {"denom": "ustake","enabled": true}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.bank.denom_metadata[0] =  {"description": "The native staking token of the stake Protocol.","denom_units": [{"denom": "ustake","exponent": 0,"aliases": ["microstake"]},{"denom": "mstake","exponent": 3,"aliases": ["millistake"]},{"denom": "stake","exponent": 6,"aliases": []}],"base": "ustake","display": "stake","name": "stake token","symbol": "stake"}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.bank.denom_metadata[1] =  {"description": "The native evm token of the stake Protocol.","denom_units": [{"denom": "astake","exponent": 0,"aliases": ["attostake"]},{"denom": "ustake","exponent": 12,"aliases": ["microstake"]},{"denom": "mstake","exponent": 15,"aliases": ["millistake"]},{"denom": "stake","exponent": 18,"aliases": []}],"base": "astake","display": "astake","name": "estake token","symbol": "astake"}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.nftadmin.authorization = {"root_admin": "cosmos1t3p2vzd7w036ahxf4kefsc9sn24pvlqpcktgg7"}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.nftmngr.nft_fee_config = {"schema_fee": {"fee_amount": "200000000ustake","fee_distributions": [{"method": "BURN","portion": 0.5},{"method": "REWARD_POOL","portion": 0.5}]}}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.nftoracle.params = {"action_request_active_duration": "120s","mint_request_active_duration": "120s","verify_request_active_duration": "120s", "action_signer_active_duration": "2592000s","sync_action_signer_active_duration": "300s"}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.nftoracle.oracle_config = {"minimum_confirmation": 4}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.staking.validator_approval.approver_address = "cosmos1t3p2vzd7w036ahxf4kefsc9sn24pvlqpcktgg7"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.gov.deposit_params.max_deposit_period = "300s"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.gov.voting_params.voting_period = "300s"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json


if [[ $1 == "pending" ]]; then
  if [[ "$OSTYPE" == "darwin"* ]]; then
      sed -i '' 's/stake/ustake/g' ${STAKE_HOME}/config/genesis.json
      sed -i '' 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' ${STAKE_HOME}/config/config.toml
      sed -i '' 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ${SIX_HOME}/config/config.toml
      sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ${SIX_HOME}/config/config.toml
      sed -i '' 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' ${STAKE_HOME}/config/config.toml
      sed -i '' 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' ${STAKE_HOME}/config/config.toml
      sed -i '' 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' ${STAKE_HOME}/config/config.toml
      sed -i '' 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' ${STAKE_HOME}/config/config.toml
      sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ${SIX_HOME}/config/config.toml
      sed -i '' 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' ${STAKE_HOME}/config/config.toml
  else
      sed -i 's/stake/ustake/g' ${STAKE_HOME}/config/genesis.json
      sed -i 's/create_empty_blocks_interval = "0s"/create_empty_blocks_interval = "30s"/g' ${STAKE_HOME}/config/config.toml
      sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ${SIX_HOME}/config/config.toml
      sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ${SIX_HOME}/config/config.toml
      sed -i 's/timeout_prevote = "1s"/timeout_prevote = "10s"/g' ${STAKE_HOME}/config/config.toml
      sed -i 's/timeout_prevote_delta = "500ms"/timeout_prevote_delta = "5s"/g' ${STAKE_HOME}/config/config.toml
      sed -i 's/timeout_precommit = "1s"/timeout_precommit = "10s"/g' ${STAKE_HOME}/config/config.toml
      sed -i 's/timeout_precommit_delta = "500ms"/timeout_precommit_delta = "5s"/g' ${STAKE_HOME}/config/config.toml
      sed -i 's/timeout_commit = "5s"/timeout_commit = "150s"/g' ${STAKE_HOME}/config/config.toml
      sed -i 's/timeout_broadcast_tx_commit = "10s"/timeout_broadcast_tx_commit = "150s"/g' ${STAKE_HOME}/config/config.toml
  fi
fi

# Allocate genesis accounts (cosmos formatted addresses)
## denom ustake
simd add-genesis-account $(simd keys show -a val1 --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 11000000000ustake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a alice --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 1000000000ustake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a bob --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 1000000000ustake --keyring-backend ${KEYRING} --home ${STAKE_HOME}
simd add-genesis-account $(simd keys show -a super-admin --keyring-backend ${KEYRING} --home ${STAKE_HOME}) 1000000000ustake --keyring-backend ${KEYRING} --home ${STAKE_HOME}

echo $KEYRING
echo $KEY
# Sign genesis transaction
simd gentx val1 1000000000ustake --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
simd collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
simd validate-genesis

if [[ $1 == "pending" ]]; then
  echo "pending mode is on, please wait for the first block committed."
fi

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
simd start --minimum-gas-prices=1.25ustake --rpc.laddr "tcp://0.0.0.0:26657"

