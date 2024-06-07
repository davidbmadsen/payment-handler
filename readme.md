# Firi kodeoppgave - CLI betalingssystem

Da jeg løste oppgaven tok jeg følgende antagelser
1. Ugyldige operasjoner må bli tatt hånd om uten at applikasjonen stopper
2. Applikasjonen skal kunne håndtere feil i inndata, og fortsette til filen er lest ferdig 
3. Ved eventuelle feil i en rad vil vi ignorere transaksjonen
4. Uttak som vil føre til negativ balanse (dvs. `available` beløp) vil ikke bli gjennomført
5. Hvis en tidligere transaksjon blir disputed og brukeren har tatt ut i mellomtiden slik at dekningen på kontoen (available-beløp) ikke kan dekke hele disputen, vil transaksjonen feile da kunde ikke skal kunne ha negativ balanse
6. Lignende, hvis det ikke er nok på hold (dvs feks. hvis overfornevnte skjer), vil resolve feile
7. Chargeback vil trekke det fulle beløpet fra hold, uavhengig av hva som er på hold

#### How to run
1. `$ go build`
2. `$ ./payment-handler -f payment-file.csv` eller `$ ./payment-handler --file payment-file.csv`

Bytt ut `payment-file.csv` med ditt filnavn
