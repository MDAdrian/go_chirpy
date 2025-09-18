# go_chirpy
connection string: psql "postgres://maniuadrian:@localhost:5432/chirpy"

## Migration
Go to sq/schema
run: 
goose postgres "postgres://maniuadrian:@localhost:5432/chirpy" up


## PSQL commands
show tables
\dt


## Packages
go get -u github.com/golang-jwt/jwt/v5