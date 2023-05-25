package main

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func main() {
	//inicializa o navegador e abre a página de login
	page := rod.New().NoDefaultDevice().MustConnect().MustPage("https://www.fazenda.sp.gov.br/folha/nova_folha/acessar_dce.asp?menu=dem&user=rs").MustWaitLoad()

	//preenche os campos de login
	page.MustElement("input[name='txt_logindce']").MustInput("6880861")
	page.MustElement("input[name='txt_senhadce']").MustInput("pj270406")

	//clica no botão de login
	page.MustElement("input[type='submit'][name='enviar']").MustClick()

	//aguardar carregar a página
	page.MustWaitLoad()

	//clica no ok do popup de abertura
	page.MustElement("input[type='button'][value='OK']").MustClick()

	//defini o primeiro link do demonstrativo de pagamento
	//linkPagina2 := fmt.Sprintf("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_listar.asp?opcao=&acao=&page=2&ano=%d&page_ano=", ano)
	page.Navigate("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_imp.asp?sq=1&tp=0&dt=12304&rb=0&rs=6880861&nro=0&tabela=atual&sit=1&dt_sit=&pv=01&opcao_pagto=visualizar&tipo_usuario=rs&opcao=abertura&acao=&ver_aviso=true&modo=imprimir")

	//aguardar 2 segundos
	time.Sleep(2 * time.Second)
	//printar a tela
	//page.MustScreenshot("teste.pdf")

	// salvar a página como PDF
	pdf, err := page.PDF(&proto.PagePrintToPDF{})
	if err != nil {
		panic(err)
	}

	// escrever o conteúdo do PDF em um arquivo
	err = pdf.Save("pagina.pdf")
	if err != nil {
		panic(err)
	}

}
