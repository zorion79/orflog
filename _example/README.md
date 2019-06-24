# zorion79/orflog

## build and try

- run `go run main.go`

## parameters

No command line parameters nedeed. You need define env variables

- `EXMPL_ORFLOG_LOGPATH` - path to logs `\\orf01\ORF\,\\orf02\ORF\`
- `EXMPL_ORFLOG_LOGSUFFIX` - log extension `.log`
- `EXMPL_ORFLOG_GOROUTINECOUNT_STRINGS` - `40`
- `EXMPL_ORFLOG_GOROUTINECOUNT_ORFRECORDS` - `10000`
- `EXMPL_ORFLOG_ORFLINE` - find string in log `SMTPSVC`

Period
- `EXMPL_ORFLOG_YEARS` - `0` years
- `EXMPL_ORFLOG_MOTHS` - `1` months
- `EXMPL_ORFLOG_DAYS` - `0` days