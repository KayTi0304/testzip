# testzip

PS C:\Users\Roy Ng\Documents\1tplus\Github\trading-ops> & ".venv/Scripts/Activate.ps1"                                                                                                                                                                                                                      
(.venv) PS C:\Users\Roy Ng\Documents\1tplus\Github\trading-ops> $env:AWS_REGION = "ap-southeast-1"
>> $env:HOST_API_SECRET = "arn:aws:secretsmanager:ap-southeast-1:285228315541:secret:local/1tplus/bnb/HostApi-xwg5gW"
>> $env:DB_CREDENTIALS_SECRET = "arn:aws:secretsmanager:ap-southeast-1:285228315541:secret:uat/1tplus/mysql-IQO4Zl"
(.venv) PS C:\Users\Roy Ng\Documents\1tplus\Github\trading-ops> go run lambda_functions/binance_futures_service/futures_signal_explorer/main.go
