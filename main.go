package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func main() {
	//inicializa o navegador e abre a página de login
	page := rod.New().NoDefaultDevice().MustConnect().MustPage("https://www.fazenda.sp.gov.br/folha/nova_folha/acessar_dce.asp?menu=dem&user=rs").MustWaitLoad()

	// Solicita o usuário e a senha ao usuário
	var usuario, senha string
	fmt.Print("Digite o usuário: ")
	fmt.Scan(&usuario)
	fmt.Print("Digite a senha: ")
	fmt.Scan(&senha)

	// Preenche os campos de login
	page.MustElement("input[name='txt_logindce']").MustInput(usuario)
	page.MustElement("input[name='txt_senhadce']").MustInput(senha)

	//clica no botão de login
	page.MustElement("input[type='submit'][name='enviar']").MustClick()

	//aguardar carregar a página
	page.MustWaitLoad()

	//clica no ok do popup de abertura
	page.MustElement("input[type='button'][value='OK']").MustClick()

	// Solicita ao usuário o mês e o ano desejados
	var mes, ano int
	fmt.Print("Digite o mês: ")
	fmt.Scan(&mes)
	fmt.Print("Digite o ano: ")
	fmt.Scan(&ano)

	// Gerar os últimos 60 arquivos a partir da data fornecida
	for i := 0; i < 60; i++ {
		// Constrói a URL com base no mês e ano fornecidos
		url := fmt.Sprintf("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_imp.asp?sq=1&tp=0&dt=1%02d%02d&rb=0&rs=6880861&nro=0&tabela=atual&sit=1&dt_sit=&pv=01&opcao_pagto=visualizar&tipo_usuario=rs&opcao=abertura&acao=&ver_aviso=true&modo=imprimir", ano, mes)

		// Navega para a página correspondente
		page.Navigate(url)

		// Aguarda 2 segundos
		time.Sleep(2 * time.Second)

		// Salva a página como PDF
		pdf, err := page.PDF(&proto.PagePrintToPDF{})
		if err != nil {
			panic(err)
		}

		// Lê o conteúdo do PDF
		content, err := ioutil.ReadAll(pdf)
		if err != nil {
			panic(err)
		}

		// Formata o nome do arquivo
		nomeArquivo := fmt.Sprintf("%02d-%02d.pdf", ano, mes)

		// Define o caminho completo para salvar o arquivo dentro da pasta do usuário
		caminhoArquivo := fmt.Sprintf("%s/%s", usuario, nomeArquivo)

		// Escreve o conteúdo do PDF em um arquivo com o novo nome e caminho
		err = ioutil.WriteFile(caminhoArquivo, content, 0644)
		if err != nil {
			panic(err)
		}

		// Decrementa o mês e o ano
		if mes == 1 {
			mes = 12
			ano--
		} else {
			mes--
		}
	}
}
