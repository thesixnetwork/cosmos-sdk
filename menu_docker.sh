default_STAKE_HOME=simapp_home
default_docker_tag=latest
node_homes=(
    simnode0
    simnode1
    simnode2
    simnode3
)
validator_keys=(
    val1
    val2
    val3
    val4
)

function setUpGenesis() {
    ## config genesis.json
    jq '.app_state.bank.params.send_enabled[0] = {"denom": "stake","enabled": true}' ./build/simnode0/config/genesis.json | sponge ./build/simnode0/config/genesis.json

    ## from stake to stake
    sed -i '' "s/stake/stake/g" ./build/simnode0/config/genesis.json

    ## staking
    jq '.app_state.staking.validator_approval.approver_address = "cosmos1t3p2vzd7w036ahxf4kefsc9sn24pvlqpcktgg7"' ./build/simnode0/config/genesis.json | sponge ./build/simnode0/config/genesis.json
    jq '.app_state.staking.params.unbonding_time = "300s"' ./build/simnode0/config/genesis.json | sponge ./build/simnode0/config/genesis.json

    ## gov
    jq '.app_state.gov.deposit_params.max_deposit_period = "300s"' ./build/simnode0/config/genesis.json | sponge ./build/simnode0/config/genesis.json
    jq '.app_state.gov.voting_params.voting_period = "300s"' ./build/simnode0/config/genesis.json | sponge ./build/simnode0/config/genesis.json
}

function setUpConfig() {
    echo "#######################################"
    echo "Setup ${STAKE_HOME} genesis..."

    if [[ ${STAKE_HOME} == "simnode0" ]]; then
        echo "simnode0"
        setUpGenesis
    else
        NODE_PEER=$(jq '.app_state.genutil.gen_txs[0].body.memo' ./build/simnode0/config/genesis.json)
        if [[ "$OSTYPE" == "darwin"* ]]; then
            ## replace NODE_PEER in config.toml to persistent_peers
            sed -i '' "s/persistent_peers = \"\"/persistent_peers = ${NODE_PEER}/g" ./build/${STAKE_HOME}/config/config.toml
        else
            sed -i "s/persistent_peers = \"\"/persistent_peers = ${NODE_PEER}/g" ./build/${STAKE_HOME}/config/config.toml
        fi
        ## replace genesis of node0 to all node
        cp ./build/simnode0/config/genesis.json ./build/${STAKE_HOME}/config/genesis.json
    fi

    # if $TYPE = 0 then ignore this step
    if [[ ${TYPE} == "1" ]]; then
        echo "Running Fast Node"
        ## replace consensus params
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/timeout_propose = \"3s\"/timeout_propose = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
            sed -i '' "s/timeout_commit = \"5s\"/timeout_commit = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
        else
            sed -i "s/timeout_propose = \"3s\"/timeout_propose = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
            sed -i "s/timeout_commit = \"5s\"/timeout_commit = \"1s\"/g" ./build/${STAKE_HOME}/config/config.toml
        fi
    else
        echo "Running Default Node"
    fi

    if [[ "$OSTYPE" == "darwin"* ]]; then
        ## replace to enalbe api
        sed -i '' '/^\[api\]$/,/^\[/ s/enable = false/enable = true/' ./build/${STAKE_HOME}/config/app.toml
        sed -i '' '/^\[api\]$/,/^[^[]/ s/^swagger = false$/swagger = true/' ./build/${STAKE_HOME}/config/app.toml
        ## replace to from 127.0.0.1 to 0.0.0.0
        sed -i '' "s/127.0.0.1/0.0.0.0/g" ./build/${STAKE_HOME}/config/config.toml

        ## replace mininum gas price
        sed -i '' "s/minimum-gas-prices = \"0stake\"/minimum-gas-prices = \"1.25stake\"/g" ./build/${STAKE_HOME}/config/app.toml
    else
        sed -i '/^\[api\]$/,/^\[/ s/enable = false/enable = true/' ./build/${STAKE_HOME}/config/app.toml
        sed -i '/^\[api\]$/,/^[^[]/ s/^swagger = false$/swagger = true/' ./build/${STAKE_HOME}/config/app.toml
        ## replace to from 127.0.0.1 to 0.0.0.0
        sed -i "s/127.0.0.1/0.0.0.0/g" ./build/${STAKE_HOME}/config/config.toml

        ## replace mininum gas price
        sed -i "s/minimum-gas-prices = \"0stake\"/minimum-gas-prices = \"1.25stake\"/g" ./build/${STAKE_HOME}/config/app.toml
    fi

    echo "Setup Genesis Success 🟢"

}

echo "#############################################"
echo "## 1. Build Docker Image                   ##"
echo "## 2. Docker Compose init chain            ##"
echo "## 3. Start chain validator                ##"
echo "## 4. Stop chain validator                 ##"
echo "## 5. Config Genesis                       ##"
echo "## 6. Reset chain validator                ##"
echo "## 7. Staking validator                    ##"
echo "## 8. Query Validator set                  ##"
echo "## 9. Setup Cosmovisor                     ##"
echo "## 10. Start Cosmovisor                    ##"
echo "#############################################"
read -p "Enter your choice: " choice
case $choice in
1)
    echo "Building Docker Image"
    read -p "Enter Docker Tag: " docker_tag
    if [ -z "$docker_tag" ]; then
        docker_tag=$default_docker_tag
    fi
    docker build . -t cosmossdk/simnode:${docker_tag}
    ;;
2)
    echo "Run init Chain validator"
    export COMMAND="init"
    docker compose -f ./docker-compose.yml up -d
    ;;
3)
    echo "Running Docker Container in Interactive Mode"
    export COMMAND="start_chain"
    docker compose -f ./docker-compose.yml up -d
    ;;
4)
    echo "Stop Docker Container"
    export COMMAND="start_chain"
    docker compose -f ./docker-compose.yml down
    ;;
5)
    echo "Config Genesis"
    read -p "Enter Node Type [0:Default, 1:Fast] : " TYPE
    if [ -z "$TYPE" ]; then
        TYPE=0
    fi
    for home in ${node_homes[@]}; do
        (
            export STAKE_HOME=${home}
            if [[ -e !./build/simnode0/config/genesis.json ]]; then
                echo "File does not exist 🖕"
            else
                setUpConfig
            fi
        ) || exit 1
    done
    ;;
6)
    echo "Reset Docker Container"
    for home in ${node_homes[@]}; do
        echo "#######################################"
        echo "Starting ${home} reset..."

        (
            export DAEMON_HOME=./build/${home}
            rm -rf $DAEMON_HOME/data
            rm -rf $DAEMON_HOME/wasm
            rm $DAEMON_HOME/config/addrbook.json
            mkdir $DAEMON_HOME/data/
            touch $DAEMON_HOME/data/priv_validator_state.json
            echo '{"height": "0", "round": 0,"step": 0}' >$DAEMON_HOME/data/priv_validator_state.json

            echo "Reset ${home} Success 🟢"
        ) || exit 1
    done
    ;;
7)
    echo "Staking Docker Container"
    read -p "Chain ID [testnet] : " CHAIN_ID
    if [ -z "$CHAIN_ID" ]; then
        CHAIN_ID="testnet"
    fi
    i=1
    amount=100000000
    # i=0
    # for val in ${validator_keys[@]}
    for val in ${validator_keys[@]:1:3}; do
        # if i=3, echo "#######################################"
        if [[ $i -eq 2 ]]; then
            echo "#######################################"
            (
                echo "Creating validators ${val}"
                echo ${node_homes[i]}
                export DAEMON_HOME=./build/${node_homes[i]}
                simd tx staking create-validator --amount 1000000stake --license-mode=true --max-license=1 --pubkey $(simd tendermint show-validator --home ./build/${node_homes[i]}) --home build/${node_homes[i]} \
                    --min-delegation 1000000 --delegation-increment 1000000 --enable-redelegation=false --moniker ${node_homes[i]} --from=${val} \
                    --commission-rate "0.1" --commission-max-rate "0.1" \
                    --commission-max-change-rate "0.1" --chain-id $CHAIN_ID \
                    --sign-mode amino-json --gas auto --gas-adjustment 1.5 --gas-prices 1.25stake --min-self-delegation 1000000 --keyring-backend test -y
                echo "Config Genesis at ${home} Success 🟢"
            ) || exit 1
        else
            echo "#######################################"
            (
                echo "Creating validators ${val}"
                echo ${node_homes[i]}
                export DAEMON_HOME=./build/${node_homes[i]}
                simd tx staking create-validator --amount="${amount}stake" --from=${val} --moniker ${node_homes[i]} \
                    --pubkey $(simd tendermint show-validator --home ./build/${node_homes[i]}) --home build/${node_homes[i]} \
                    --keyring-backend test --commission-rate 0.1 --commission-max-rate 0.5 --commission-max-change-rate 0.1  --max-license 100\
                    --min-self-delegation 1000000 --node http://0.0.0.0:26662 -y --min-delegation 1000000 --delegation-increment 1000000 \
                    --chain-id $CHAIN_ID --gas auto --gas-adjustment 1.5 --gas-prices 1.25stake -y
                echo "Config Genesis at ${home} Success 🟢"
            ) || exit 1
        fi
        i=$((i + 1))
    done
    ;;
8)
    echo "Query Validator set"
    simd q tendermint-validator-set --home ./build/simnode0
    ;;
9)
    echo "Set up Cosmovisor"
    export COMMAND="cosmovisor_setup"
    docker compose -f ./docker-compose.yml up -d
    ;;
10)
    echo "Cosmovisor start"
    export COMMAND="cosmovisor_start"
    docker compose -f ./docker-compose.yml up -d
    ;;
*)
    echo "Invalid Choice"
    ;;
esac
