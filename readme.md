# Betalingssystem - Firi kodeoppgave

Da jeg løste oppgaven tok jeg følgende antagelser
1. Applikasjonen skal kunne håndtere feil inndata, og fortsette til filen er lest ferdig
2. Ved eventuelle feil vil vi ignorere transaksjonen
3. Ugyldige operasjoner må bli tatt hånd om uten at applikasjonen stopper
4. Uttak som vil føre til negativ balanse (available) vil ikke bli gjennomført
5. Hvis en bruker har tatt ut et beløp etter en transaksjon som har blitt disputet, vil fortsatt det fulle beløpet bli satt på hold, selv om det vil tilsi et negativt hold-beløp

#### How to run
1. `$ go build`
2. `$ ./payment-handler -f payment-file.csv` eller `$ ./payment-handler --file payment-file.csv`

Bytt ut `payment-file.csv` med ditt filnavn
