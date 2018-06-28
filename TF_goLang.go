// ----------------------------------------------------------------------------------
// Leitores e Escritores em Go segundo modelo CSP abaixo
// Por Prof. Dotti - Escola Politecnica - PUCRS
// Abaixo o modelo, a seguir a implementacao em Go, com cada parte do modelo mapeada ----------------------------------------------------------------------------------
// --  Leitores Escritores
// --  Dotti

// QL = 2 -- Quantidade de Leitores
// QE = 2 -- Quantidade de Escritores

package main

const QL = 4 // Quantidade de leitores
const QE = 3 // Quantidade de escritores

//
// IDL = {1 .. QL}
// IDE = {1 .. QE}
//            -------------- isto nao foi mapeado
// Valor = IDE
//            -------------- valor vai ser int
// -- canais
//
// channel leitorEntra,leitorSai: IDL
// channel leitorLendo:IDL.Valor
//
// channel escritorEntra,escritorSai: IDE
// channel escritorEscreve:IDE.Valor


func main() {
	// canais sincronizantes (tamanho 0) para inicio e fim de leitura e escrita
	var leitorEntra [QL]chan struct{} // inicL[i] - leitor i faz inic
	var leitorSai [QL]chan struct{}  // fimL[i] - leitor i faz fim

	var escritorEntra [QE]chan struct{} // inicL[i] - leitor i faz inic
	var escritorSai [QE]chan struct{}  // fimL[i] - leitor i faz fim

	var leitorLendo chan int // acesso ao dado
	var escritorEscreve chan int

	for i := 1; i < QL; i++ {
		leitorEntra[i] = make(chan struct{})
		leitorSai[i] = make(chan struct{})
	}
	for i := 1; i < QE; i++ {
		escritorEntra[i] = make(chan struct{})
		escritorSai[i] = make(chan struct{})
	}
	leitorLendo = make(chan int)
	escritorEscreve = make(chan int)

	//
	// -- composicao
	//
	// RW = (Dado ||| Controle) [| union(IfEscritores,IfLeitores)|] (Leitores ||| Escritores)
	//
	// --

	// processo RW nao mapeado.   aqui simplesmente lancamos cada
	// elemento do interleaving de RW
	// note que a definicao de interface ee implicita em Go

	go Dado(0, leitorLendo, escritorEscreve)
	go Leitores(leitorEntra, leitorSai, leitorLendo)
	go Escritores(escritorEntra, escritorSai, escritorEscreve)
	go Controle(leitorEntra, leitorSai, escritorEntra, escritorSai)

	var bloq chan int
	<-bloq // bloqueia para main nao acabar

}

// -- Dado
//
// Dado = D(1)
// D(v) =     le?l!v -> D(v)
//        []  escreve?e?nv -> D(nv)

func Dado(inic int, leitorLendo chan int, escritorEscreve chan int) {
	var valor int
	valor = inic
	for {
		select {
		case leitorLendo <- valor:
		case valor = <-escritorEscreve:
		}

	}
}

// -- Leitores
//
// Leitor(n) = leitorEntra!n -> leitorLendo.n?v -> leitorSai!n -> Leitor(n)
// Leitores = ||| n:IDL @ Leitor(n)
// AcoesLeitores = { leitorEntra.n, leitorLendo.n.v, leitorSai.n | n: IDL , v: Valor}

func Leitor(id int,
	inicio chan struct{},
	fim chan struct{},
	le chan int) {
	var valor int
	for {
		<-inicio
		valor = <-le
		println("Leitura do Leitor ", id, " ", valor)
		<-fim
	}
}

func Leitores(leitorEntra [QL]chan struct{}, leitorSai [QL]chan struct{}, leitorLendo chan int) {
	for i := 1; i < QL; i++ {
		go Leitor(i, leitorEntra[i], leitorSai[i], leitorLendo)
	}
}

// -- Escritores
//
// Escritor(n) = escritorEntra!n -> escritorEscreve.n!n-> escritorSai!n -> Escritor(n)
// Escritores = ||| n:IDE @ Escritor(n)
// AcoesEscritores = { escritorEntra.n, escritorEscreve.n.v, escritorSai.n | n: IDE , v: Valor}
//

func Escritor(id int,
	inicio chan struct{},
	fim chan struct{},
	escreve chan int) {
	var valor int
	for {
		<-inicio
		valor = <-escreve
		println("Escrita do Escritor ", id, " ", valor)
		<-fim
	}
}

func Escritores(escritorEntra [QE]chan struct{}, escritorSai [QE]chan struct{}, escritorEscreve chan int) {
	for i := 1; i < QE; i++ {
		go Escritor(i, escritorEntra[i], escritorSai[i], escritorEscreve)
	}
}

// Controle = C(0)
// C(s) =
//     if s==0
//     then ( inicL?l -> C(1)
// 	       []
//   		   inicE?e -> fimE.e -> C(s) )
//     else if s<NL
// 	     then ( inicL?l -> C(s+1)
// 		        []
// 		        fimL?l -> C(s-1)  )
// 		 else fimL?l -> C(s-1)

func Controle(leitorEntra [QL]chan struct{}, leitorSai [QL]chan struct{},
	escritorEntra [QE]chan struct{}, escritorSai [QE]chan struct{}) {
	var s int
//	var iL // indice Leitor/Escritor
	s = 0

	for {
		if s == 0 {
			select { // infelizmente esta construcao ee fixa para o numero de leitore e escritores.
			// temos que descobrir como parametrizar isso em Go

			case leitorEntra[1] <- struct{}{}:
				s++
			case leitorEntra[2] <- struct{}{}:
				s++
			case escritorEntra[1] <- struct{}{}:
				escritorSai[1] <- struct{}{}
			case escritorEntra[2] <- struct{}{}:
				escritorSai[2] <- struct{}{}
			}
		} else if (s > 0) && (s < QL) {
				select {
				case leitorEntra[1] <- struct{}{}:
					s++
				case leitorEntra[2] <- struct{}{}:
					s++
				case leitorSai[1] <- struct{}{}:
					s--
				case leitorSai[2] <- struct{}{}:
					s--
				}
		} else {
				select {
				case leitorSai[1] <- struct{}{}:
					s--
				case leitorSai[2] <- struct{}{}:
					s--
				}
		}
	}
}
