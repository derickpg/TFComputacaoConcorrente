// ----------------------------------------------------------------------------------
// Leitores e Escritores em Go segundo modelo CSP abaixo
// Por Prof. Dotti - Escola Politecnica - PUCRS
// Abaixo o modelo, a seguir a implementacao em Go, com cada parte do modelo mapeada ----------------------------------------------------------------------------------
// --  Leitores Escritores
// --  Dotti
//
// NL = 2     -- numero de leitores
// NE = 2     -- numero de escritores
//
// IdLeitores =   {1 .. NL}
// IdEscritores = {1 .. NE}
//
// Valor = IdEscritores
//
// -- canais
//
// channel inicL,fimL: IdLeitores
// channel inicE,fimE: IdEscritores
//
// channel le:IdLeitores.Valor
// channel escreve:IdEscritores.Valor
//
// -- leitores e escritores
//
// Leitor(i) = inicL!i -> le.i?v -> fimL!i -> Leitor(i)
// Leitores = ||| i:IdLeitores @ Leitor(i)
// IfLeitores = { inicL.i, le.i.v, fimL.i | i: IdLeitores , v: Valor}
//
// Escritor(i) = inicE!i -> escreve.i!i-> fimE!i -> Escritor(i)
// Escritores = ||| i:IdEscritores @ Escritor(i)
// IfEscritores = { inicE.i, escreve.i.v, fimE.i | i: IdEscritores , v: Valor}
//
// -- Recurso
//
// Dado = D(1)
// D(v) =     le?l!v -> D(v)
//        []  escreve?e?nv -> D(nv)
//
// -- composicao
//
// RW = (Dado ||| Controle) [| union(IfEscritores,IfLeitores)|] (Leitores ||| Escritores)
//
// --
//
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

// ----------------------------------------------------------------------------------
// Agora o codigo Go para cada parte CSP
// ----------------------------------------------------------------------------------

// NL = 2     -- numero de leitores
// NE = 2     -- numero de escritores

package main

const NL = 4 // numero de leitores
const NE = 3 // numero de escritores

//
// IdLeitores =   {1 .. NL}
// IdEscritores = {1 .. NE}
//            -------------- isto nao foi mapeado
// Valor = IdEscritores
//            -------------- valor vai ser int
// -- canais
//
// channel inicL,fimL: IdLeitores
// channel inicE,fimE: IdEscritores
//
// channel le:IdLeitores.Valor
// channel escreve:IdEscritores.Valor
//

func main() {
	// canais sincronizantes (tamanho 0) para inicio e fim de leitura e escrita
	var inicL [NL]chan struct{} // inicL[i] - leitor i faz inic
	var fimL [NL]chan struct{}  // fimL[i] - leitor i faz fim

	var inicE [NE]chan struct{} // inicL[i] - leitor i faz inic
	var fimE [NE]chan struct{}  // fimL[i] - leitor i faz fim

	var le chan int // acesso ao dado
	var escreve chan int

	for i := 1; i < NL; i++ {
		inicL[i] = make(chan struct{})
		fimL[i] = make(chan struct{})
	}
	for i := 1; i < NE; i++ {
		inicE[i] = make(chan struct{})
		fimE[i] = make(chan struct{})
	}
	le = make(chan int)
	escreve = make(chan int)

	//
	// -- composicao
	//
	// RW = (Dado ||| Controle) [| union(IfEscritores,IfLeitores)|] (Leitores ||| Escritores)
	//
	// --

	// processo RW nao mapeado.   aqui simplesmente lancamos cada
	// elemento do interleaving de RW
	// note que a definicao de interface ee implicita em Go

	go Dado(0, le, escreve)
	go Leitores(inicL, fimL, le)
	go Escritores(inicE, fimE, escreve)
	go Controle(inicL, fimL, inicE, fimE)

	var bloq chan int
	<-bloq // bloqueia para main nao acabar

}

// -- Dado
//
// Dado = D(1)
// D(v) =     le?l!v -> D(v)
//        []  escreve?e?nv -> D(nv)

func Dado(inic int, le chan int, escreve chan int) {
	var valor int
	valor = inic
	for {
		select {
		case le <- valor:
		case valor = <-escreve:
		}

	}
}

// -- Leitores
//
// Leitor(i) = inicL!i -> le.i?v -> fimL!i -> Leitor(i)
// Leitores = ||| i:IdLeitores @ Leitor(i)
// IfLeitores = { inicL.i, le.i.v, fimL.i | i: IdLeitores , v: Valor}

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

func Leitores(inicL [NL]chan struct{}, fimL [NL]chan struct{}, le chan int) {
	for i := 1; i < NL; i++ {
		go Leitor(i, inicL[i], fimL[i], le)
	}
}

// -- Escritores
//
// Escritor(i) = inicE!i -> escreve.i!i-> fimE!i -> Escritor(i)
// Escritores = ||| i:IdEscritores @ Escritor(i)
// IfEscritores = { inicE.i, escreve.i.v, fimE.i | i: IdEscritores , v: Valor}
//

//   POR FAZER!!!!

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

func Controle(inicL [NL]chan struct{}, fimL [NL]chan struct{},
	inicE [NE]chan struct{}, fimE [NE]chan struct{}) {
	var s int
	s = 0

	for {
		if s == 0 {
			select { // infelizmente esta construcao ee fixa para o numero de leitore e escritores.
			// temos que descobrir como parametrizar isso em Go
			case inicL[1] <- struct{}{}:
				s++
			case inicL[2] <- struct{}{}:
				s++
			case inicE[1] <- struct{}{}:
				fimE[1] <- struct{}{}
			case inicE[2] <- struct{}{}:
				fimE[2] <- struct{}{}
			}
		}
		// RESTO POR FAZER!!!
		if (s > 0) && (s < NL) {

		}
	}
}
