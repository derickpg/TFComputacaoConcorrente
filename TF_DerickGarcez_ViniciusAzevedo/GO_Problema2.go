// -----------------------------------------------------------------------------
// Trabalho Modelos para Computacao Concorrente
// Leitores e Escritores (problema 2) - modelo GO
// Alunos: Derick Garcez
//        Vinicius Azevedo
// Escola Politecnica - PUCRS
// Desenvolvido a partir do exemplo disponibilizado pelo professor
// -----------------------------------------------------------------------------

package main

const QL = 4 // numero de leitores
const QE = 3 // numero de escritores

func main() {
	// declarar lista de canais para inicio e fim de leitura e escrita
	var leitorEntra [QL]chan struct{}
	var leitorSai [QL]chan struct{}
	var escritorEntra [QE]chan struct{}
	var escritorSai [QE]chan struct{}

	// declaracao de canais de acesso ao dado
	var leitorLendo chan int
	var escritorEscreve chan int

	// para cada leitor criar canal de entrada e saida
	for i := 1; i < QL; i++ {
		leitorEntra[i] = make(chan struct{})
		leitorSai[i] = make(chan struct{})
	}

	// para cada escritor criar canal de entrada e saida
	for i := 1; i < QE; i++ {
		escritorEntra[i] = make(chan struct{})
		escritorSai[i] = make(chan struct{})
	}

	//inicializa canais de leitura e escrita (dado = 0)
	leitorLendo = make(chan int)
	escritorEscreve = make(chan int)

	//inicia goroutine de todos processos em paralelo
	go Dado(0, leitorLendo, escritorEscreve)
	go Leitores(leitorEntra, leitorSai, leitorLendo)
	go Escritores(escritorEntra, escritorSai, escritorEscreve)
	go Controle(leitorEntra, leitorSai, escritorEntra, escritorSai)

	// ? ? ?
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
		case valor = <- escritorEscreve:
		}

	}
}

// coloca todos leitores executando simultaneamente
//com canais de entrada, saida e de leitura
func Leitores(leitorEntra [QL]chan struct{}, leitorSai [QL]chan struct{}, leitorLendo chan int) {
	for i := 1; i < QL; i++ {
		go Leitor(i, leitorEntra[i], leitorSai[i], leitorLendo)
	}
}
func Leitor(id int,
	inicio chan struct{},
	fim chan struct{},
	leitorLendo chan int) {
	var valor int
	for {
		<-inicio
		valor = <-leitorLendo
		println("Leitura do Leitor ", id, " ", valor)
		<-fim
	}
}

// coloca todos escritores juntos para executar simultaneamente
//com canais de entrada, saida e de escrita
func Escritores(escritorEntra [QE]chan struct{}, escritorSai [QE]chan struct{}, escritorEscreve chan int) {
	for i := 1; i < QE; i++ {
		go Leitor(i, escritorEntra[i], escritorSai[i], escritorEscreve)
	}
}
func Escritor(id int,
	inicio chan struct{},
	fim chan struct{},
	escritorEscreve chan int) {
	var valor int
	for {
		<-inicio
		valor = <- escritorEscreve
		println("Escritor ", id, " escrevendo ", valor)
		<-fim
	}
}

//esse e o juiz que vai controlar o acesso de leitores e escritores
//
func Controle(leitorEntra [QL]chan struct{}, leitorSai [QL]chan struct{},
	escritorEntra [QE]chan struct{}, escritorSai [QE]chan struct{}) {
	var s, y, lx, ex int
	s, y, lx, ex = 0, 0, 1, 1

	for {
		if s == 0 && y == 0 {
			//se nao tem ninguem acessando, ESPERA UMA ACAO:
			select {
				case leitorEntra[lx] <- struct{}{}: // entra leitor lx
					s++
					lx++
				case escritorEntra[ex] <- struct{}{}: // escritor acessa
					y++
					ex++
			}
		} else if y == 0 && s < QL {
			//se tem pelo menos um leitor acessando, mas nenhum escritor
			select {
				case escritorEntra[ex] <- struct{}{}: // escritor acessa
					y++
					ex++
				case leitorEntra[lx] <- struct{}{}: // pode entrar mais leitores
					s++
					lx++
				case leitorSai[lx] <- struct{}{}: // ou sair um leitor
					s--
					lx--
			}
		} else if s == 0 && y < QE {
			// se apenas tem escritores acessando
			select{
				case escritorEntra[ex] <- struct{}{}: // escritor acessa
					y++
					ex++
				case escritorSai[ex] <- struct{}{}: // escritor sai
					y--
					ex--
			}
		} else if y >= QE {
			//neste caso, todos escritores estao acessando
			select {
				case escritorSai[ex] <- struct{}{}: // unica opcao é sair escritor
					y--
					ex--
			}
		} else {
			//neste caso, todos leitores estao acessando
			select {
				case leitorSai[lx] <- struct{}{}: // unica opcao é sair leitor
					s--
					lx--
			}
		}
	}
}
