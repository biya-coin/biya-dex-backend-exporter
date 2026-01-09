## explorer
    - 指标名称： biya_block_height 
      指标描述： 当前区块高度
      获取的地址： curl --request GET  --url https://prv.explorer.biya.io/api/v1/block/latest
      取值： .data.data.[0].height
    - 指标名称: biya_tx_24h_total
      指标描述： 24h交易数
      获取的地址：  curl --request GET --url https://prv.explorer.biya.io/api/v1/transaction/stats
      取值： .data.count_24h
    - 指标名称：biya_tps_current
      指标描述：当前网络tps
      获取的地址：  curl --request GET --url https://prv.explorer.biya.io/api/v1/transaction/stats
      取值： .data.tps
    - 指标名称：biya_block_time_seconds
      指标描述：平均出块时间
      获取的地址：  curl --request GET --url https://prv.explorer.biya.io/api/v1/transaction/stats
      取值： .data.avg_block_time
    - 指标名称：biya_validators_active
      指标描述：活跃节点数
      获取的地址：  curl --request GET   --url https://prv.stake.biya.io/stake/validators
      取值： .len(validators)
    - 指标名称：biya_active_addresses_24h
      指标描述：24h活跃地址数
      获取的地址：  curl --request GET --url https://prv.explorer.biya.io/api/v1/transaction/stats
      取值： .data.active_addresses_24h
    - 指标名称：biya_gas_price
      指标描述：当前Gas费
      获取的地址：  curl --request GET --url https://prv.explorer.biya.io/api/v1/block/gas-utilization
      取值： .data.gas_price
    - 指标名称：biya_gas_utilization
      指标描述：Gas利用率
      获取的地址：  curl --request GET --url https://prv.explorer.biya.io/api/v1/block/gas-utilization
      取值： .data.gas_utilization

## chain node
    - 指标名称：biya_mempool_capacity 
      指标描述：交易池的大小 
      获取的地址：默认配置5000
      取值：5000
    - 指标名称：biya_mempool_size
      指标描述：当前交易池中pending的交易数 
      获取的地址：curl -s http://45.249.245.183:26657/num_unconfirmed_txs
      取值：.result.total







    