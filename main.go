package main

import (
	"time"

	"github.com/go-rod/rod"
)

func main() {
	// Inicializa o navegador e abre a página de login
	page := rod.New().NoDefaultDevice().MustConnect().MustPage("https://www.fazenda.sp.gov.br/folha/nova_folha/acessar_dce.asp?menu=dem&user=rs").MustWaitLoad()

	// preenche os campos de login
	page.MustElement("input[name='txt_logindce']").MustInput("6880861")
	page.MustElement("input[name='txt_senhadce']").MustInput("pj270406")

	// clica no botão de login
	page.MustElement("input[type='submit'][name='enviar']").MustClick()

	// aguardar carregar a página
	page.MustWaitLoad()

	// clica no ok do popup de abertura
	page.MustElement("input[type='button'][value='OK']").MustClick()

	// obter a data atual menos 5 anos
	// dataAnosAtras := time.Now().AddDate(-5, 0, 0)

	// obter o ano e o mês correspondentes
	// ano := dataAnosAtras.Year()
	// mes := dataAnosAtras.Month()

	// navegar para a página desejada
	page.MustNavigate("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_imp.asp?sq=1&tp=0&dt=12304&rb=0&rs=6880861&nro=0&tabela=atual&sit=1&dt_sit=&pv=01&opcao_pagto=visualizar&tipo_usuario=rs&opcao=abertura&acao=&ver_aviso=true&modo=imprimir")

	// Aguarda 5 segundos
	time.Sleep(5 * time.Second)

	// tira um screenshot após o login (exemplo)
	page.MustScreenshot("teste.png")
}
