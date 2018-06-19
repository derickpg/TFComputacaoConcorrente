// -----------------------------------------
// Este programa apresenta definicoes de semaforos em go.
// Posteriormente usa estas para definir monitores usando semaforos,
// de acordo com a literatura
//
// Apresenta entao programa exemplo usando monitor.
//
// Disciplina MCC  - PUCRS - Escola Politecnica
// Fernando Dotti
// -----------------------------------------

package main

import (
	"fmt"
	// "sync"
	//"time"
)

// -----------------------------------------
// Definição de Semáforo -------------------
// -----------------------------------------

type Semaphore struct {
	inc, dec chan struct{}
	val      int
}

func NewSemaphore(v int) *Semaphore {
	s := &Semaphore{
		inc: make(chan struct{}),
		dec: make(chan struct{}),
		val: v}

	go func() {
		for {
			if s.val == 0 {
				<-s.inc
				s.val++
			}
			if s.val > 0 {
				select {
				case <-s.inc:
					s.val++
				case <-s.dec:
					s.val--
				}
			}
		}
	}()
	return s
}

func (s *Semaphore) Signal() {
	s.inc <- struct{}{}
}

func (s *Semaphore) Wait() {
	s.dec <- struct{}{}
}

// -----------------------------------------
// Fim Definição de Semáforo ---------------
// -----------------------------------------

// Usa semáforo para SC acessando shared

type shared struct{ v int }

func useSC(sharedArea *shared, s *Semaphore, st string) {
	for {
		s.Wait()
		sharedArea.v++
		s.Signal()
		fmt.Print(sharedArea.v, st)
	}
}

func procsEmExMutua() {
	fmt.Println("inicio")
	s := NewSemaphore(0)
	sh := &shared{v: 0}
	go useSC(sh, s, "a")
	go useSC(sh, s, "b")
	s.Signal()
}

// -----------------------------------------
// Definição de Monitor  -------------------
// -----------------------------------------

// -----------------------------------------
// Observe o uso de canal como estrutura basica de sincronizacao para
// implementar semaforo, acima, em Go.   Uma vez que semaforo existe, a definicao
// de monitor nao usa mais canal, e sim o semaforo.
// Ou seja, esta implementacao de monitor com semaforo ee igual (ok... adaptada) aa
// implementacao de monitor com semaforo em C provida pelo professor
// na pagina.
// -----------------------------------------

//  ------------------------------------------------------
//  ------------------------------------------------------
//  estruturas genericas de monitores

type Monitor struct {
	mutex      *Semaphore // garante exclusão mutua do monitor
	next       *Semaphore // bloqueia thread que sinaliza em favor de outra - vide signal de condition
	next_count int        // conta threads em next, que podem prossegir
}

func initMonitor() *Monitor {
	m := &Monitor{
		mutex:      NewSemaphore(1),
		next:       NewSemaphore(0),
		next_count: 0}

	return m
}

//  procedimentos do monitor

func (m *Monitor) monitorEntry() {
	m.mutex.Wait() // entrada no monitor ee so passar pelo mutex
}

func (m *Monitor) monitorExit() {
	if m.next_count > 0 { // libera uma thread que ja esteve no monitor, senao libera mutex
		m.next.Signal()
	} else {
		m.mutex.Signal()
	}
}

//  ------------------------------------------------------
//  ------------------------------------------------------
//  estruturas genericas de  variaveis condicao

type Condition struct {
	s     *Semaphore // semaforo para bloquear na condicao
	count int        // contador de bloqueados
	m     *Monitor   // monitor associado aa condicao - quando bloqueia na condicao libera o monitor (next ou mutex)
	name  string
}

func initCondition(n string, m1 *Monitor) *Condition {
	c := &Condition{
		s:     NewSemaphore(0), // 0 inicia bloqueando
		count: 0,               // contadores de bloqueados nesta condicao
		m:     m1,              // o monitor associado
		name:  n}

	return c
}

//  procedimentos de variaveis condicao

func (c *Condition) condWait() {
	// fmt.Println("                                           wait  ", c.name)
	c.count++         // mais uma thread vai bloquear aqui nesta condition
	c.m.monitorExit() // libera o monitor associado aa condition
	c.s.Wait()        // bloqueia !!     fica aqui ate alguem dar signal  !!
	c.count--         // esta linha é executada depois de alguem ter dado signal, entao um bloqueado a menos
}

func (c *Condition) condSignal() {
	if c.count > 0 { // tem alguem para sinalizar ?    se nao tem entao nao faz nada, signal nao tem efeito!
		c.m.next_count++ // opa, tem alguem para sinalizar, entao esta thread se bloqueia em favor da sinalizada
		c.s.Signal()
		c.m.next.Wait()
		c.m.next_count-- // foi desbloqueada (veja monitorExit), aqui desbloqueou, decrementa
	}
}

//  ------------------------------------------------------
//  ------------------------------------------------------

//  ------------------------------------------------------
//  ------------------------------------------------------
//  o monitor EXEMPLO

const N = 10

type monitorBufferCircular struct {
	// sincronizacao
	m     *Monitor
	cheio *Condition
	vazio *Condition
	// dados do problema
	buffer [N]int
	in     int
	out    int
	// outros
	entraram int
	sairam   int
}

func monitorExampleInit() *monitorBufferCircular {
	mon := initMonitor()
	mbc := &monitorBufferCircular{

		m:     mon,
		cheio: initCondition("cheio", mon),
		vazio: initCondition("vazio", mon),

		in:       0,
		out:      0,
		entraram: 0,
		sairam:   0,
	}

	return mbc
}

func isFull(i int, o int) bool {
	return (((i + 1) % N) == o) // condicao de buffer cheio
}

func isEmpty(i int, o int) bool {
	return (i == o) // condicao de buffer vazio
}

func (mbc *monitorBufferCircular) insere(x int) {
	mbc.m.monitorEntry()         // entrada no monitor
	if isFull(mbc.in, mbc.out) { // se buffer cheio ?
		mbc.cheio.condWait()
	} //     entao espera nao estar mais cheio
	mbc.buffer[mbc.in] = x    // agora tem espaco, produz novo elemento
	mbc.in = (mbc.in + 1) % N // ajusta marcador de entrada

	mbc.entraram++ // contabiliza para verificar

	mbc.vazio.condSignal() // como produziu, se alguem espera para consumir pois estava vazio, pode ser acordada!!
	// se nao houver ninguem esperando, signal é vazio, nao afeta!!
	mbc.m.monitorExit() // sai do monitor
}

func (mbc *monitorBufferCircular) retira() int {
	mbc.m.monitorEntry()          // entrada no monitor
	if isEmpty(mbc.in, mbc.out) { // se buffer vazio ?
		mbc.vazio.condWait()
	} //     entao espera ter algo no buffer para poder consumir
	c := mbc.buffer[mbc.out]    // agora tem algo!  pega o item
	mbc.out = (mbc.out + 1) % N // ajusta marcador de saida

	mbc.sairam++ // contabiliza para verificar

	mbc.cheio.condSignal() // como consumiu, se alguem espera para produzir pois estava cheio, pode ser acordada!!
	mbc.m.monitorExit()    // se nao houver ninguem esperando, signal é vazio, nao afeta !!
	return c               // retorna o item consumido
}

//  ------------------------------------------------------
//  ------------------------------------------------------

//  ------------------------------------------------------
// -------------------------------------------------------
// USA monitor EXEMPLO

func consumidor(mbc *monitorBufferCircular, fin chan struct{}) {
	for {
		fmt.Println("                retira ", mbc.retira())
	}
	fin <- struct{}{}
}

func produtor(mbc *monitorBufferCircular, fin chan struct{}) {
	i := 0
	for {
		mbc.insere(i)
		fmt.Println("ins: ", i, "    in: ", mbc.in)
		i++
	}
	fin <- struct{}{}
}

//  ------------------------------------------------------
//  ------------------------------------------------------

func main() {
	fin := make(chan struct{})
	mbc := monitorExampleInit()
	go produtor(mbc, fin)
	go produtor(mbc, fin)
	go produtor(mbc, fin)
	go consumidor(mbc, fin)
	go produtor(mbc, fin)
	go produtor(mbc, fin)
	<-fin
}
