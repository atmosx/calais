# Calais
A [ledger](https://ledger-cli.org/) companion tool, to update [commodities and currencies](https://ledger-cli.org/doc/ledger3.html#Commodities-and-Currencies) values. Calais writes the latest price for each commodity or stock into the standard `prices.db` file used by ledger.


# Build

```bash
make build
```

# Configure

Calais requires a [marketstack](https://marketstack.com/) and [fixer](https://fixer.io/) accounts. Populate the configuration files with the API keys:

```yaml
# stock pricing
marketstack:
  key: "YOUR_MARKETSTACK_KEY"
  stocks:
    - AAPL
    - MSFT

fixer:
  key: "YOUR_FIXER_KEY"
  pairs:
    - { from: "EUR", to: "USD" }
    - { from: "GBP", to: "USD" }

ledger:
   price_db: "/tmp/prices.db"
```

# How to setup and use

```bash

$ echo $LEDGER_PRICE_DB
/Users/atma/.prices.db

$ cat ~/.calais/config.yaml
# stock pricing
marketstack:
  key: "<marketstack-api-key>"
  stocks:
    - TITC.AT
    - SXR8.DE

fixer:
  key: "<fixer.io-api-key>"
  pairs:
    - { from: "EUR", to: "USD" }

ledger:
   price_db: /Users/atma/.prices.db

$ calais -c ~/.calais/config.yaml
INFO[0000] wrote stock price                             date="2025-09-18 00:00:00 +0000 +0000" price=36.2 symbol=TITC.AT
INFO[0001] wrote stock price                             date="2025-09-17 00:00:00 +0000 +0000" price=595.22 symbol=SXR8.DE
INFO[0001] wrote currency price                          date="2025-09-19 08:29:07 +0300 EEST" pair=EUR/USD rate=1.17755

$ tail -n 3 ~/.prices.db
P 2025/09/18 00:00:00 TITC.AT €36.20
P 2025/09/17 00:00:00 SXR8.DE €595.22
P 2025/09/19 08:29:07 EUR $1.177550
```

## Trivia
Calais is one of the [Boreads](https://en.wikipedia.org/wiki/Boreads), sons of the North Wind (Boreas). Hailing from Thrace, Calais and his brother Zetes sailed with the Argonauts a long long time ago in a world far far away and earned fame for rescuing Phineus from the harpies.
