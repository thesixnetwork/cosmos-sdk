default_home=simapp_home
STAKE_HOME=$1
if [ -z "$STAKE_HOME" ]; then
    STAKE_HOME=$default_home
fi

## funtion setup_genesis
function setupGenesis() {
    NODE_PEER=$(jq '.app_state.genutil.gen_txs[0].body.memo' ./build/simnode0/config/genesis.json)

    if [[ "$OSTYPE" == "darwin"* ]]; then
        ## replace NODE_PEER in config.toml to persistent_peers
        sed -i '' "s/persistent_peers = \"\"/persistent_peers = ${NODE_PEER}/g" ./build/${STAKE_HOME}/config/config.toml
        ## replace mininum gas price
        sed -i '' "s/minimum-gas-prices = \"0stake\"/minimum-gas-prices = \"1.25stake\"/g" ./build/${STAKE_HOME}/config/app.toml
        ## replace to enalbe api
        sed -i '' '/^\[api\]$/,/^\[/ s/enable = false/enable = true/' ./build/${STAKE_HOME}/config/app.toml
        sed -i '' '/^\[api\]$/,/^[^[]/ s/^swagger = false$/swagger = true/' ./build/${STAKE_HOME}/config/app.toml

        ## replace to from 127.0.0.1 to 0.0.0.0
        sed -i '' "s/127.0.0.1/0.0.0.0/g" ./build/${STAKE_HOME}/config/config.toml
        ## replace consensus params
        sed -i '' "s/timeout_propose = \"3s\"/timeout_propose = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
        sed -i '' "s/timeout_commit = \"5s\"/timeout_commit = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
           ## from stake to stake
        sed -i '' "s/stake/stake/g" ./build/${STAKE_HOME}/config/genesis.json
    else 
        ## replace NODE_PEER in config.toml to persistent_peers
        sed -i "s/persistent_peers = \"\"/persistent_peers = ${NODE_PEER}/g" ./build/${STAKE_HOME}/config/config.toml
        ## replace mininum gas price
        sed -i "s/minimum-gas-prices = \"0stake\"/minimum-gas-prices = \"1.25stake\"/g" ./build/${STAKE_HOME}/config/app.toml
         ## replace to enalbe api
        sed -i '/^\[api\]$/,/^\[/ s/enable = false/enable = true/' ./build/${STAKE_HOME}/config/app.toml
        sed -i '/^\[api\]$/,/^[^[]/ s/^swagger = false$/swagger = true/' ./build/${STAKE_HOME}/config/app.toml
        ## replace to from 127.0.0.1 to 0.0.0.0
        sed -i "s/127.0.0.1/0.0.0.0/g" ./build/${STAKE_HOME}/config/config.toml
        ## replace consensus params
        sed -i "s/timeout_propose = \"3s\"/timeout_propose = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
        sed -i "s/timeout_commit = \"5s\"/timeout_commit = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
    fi 

    ## config genesis.json

    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="stake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="stake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="stake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["inflation"]["params"]["mint_denom"]="stake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.bank.params.send_enabled[0] = {"denom": "stake","enabled": true}' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.staking.validator_approval.approver_address = "cosmos1t3p2vzd7w036ahxf4kefsc9sn24pvlqpcktgg7"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.gov.deposit_params.max_deposit_period = "300s"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json
    cat ${STAKE_HOME}/config/genesis.json | jq '.app_state.gov.voting_params.voting_period = "300s"' > ${STAKE_HOME}/config/tmp_genesis.json && mv ${STAKE_HOME}/config/tmp_genesis.json ${STAKE_HOME}/config/genesis.json

    echo "Setup Genesis Success 🟢"

}

if [[ ! -e ./build/simnode0/config/genesis.json ]]; then
    echo "File does not exist 🖕"
else
    setupGenesis
fi