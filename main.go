package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func main() {
	// Inicializa o navegador e abre a página de login
	page := rod.New().NoDefaultDevice().MustConnect().MustPage("https://www.fazenda.sp.gov.br/folha/nova_folha/acessar_dce.asp?menu=dem&user=rs").MustWaitLoad()

	// Solicita o ling e senhacls
	var usuario, senha string
	fmt.Print("Digite o usuário: ")
	fmt.Scan(&usuario)
	fmt.Print("Digite a senha: ")
	fmt.Scan(&senha)

	// Preenche os campos de login
	page.MustElement("input[name='txt_logindce']").MustInput(usuario)
	page.MustElement("input[name='txt_senhadce']").MustInput(senha)

	// Clica no botão de login
	page.MustElement("input[type='submit'][name='enviar']").MustClick()

	// Aguardar carregar a página
	page.MustWaitLoad()

	// Clica no botão OK do popup de abertura
	page.MustElement("input[type='button'][value='OK']").MustClick()

	// Solicita ao usuário o mês e o ano desejados
	var mes, ano int
	fmt.Print("Digite o mês: ")
	fmt.Scan(&mes)
	fmt.Print("Digite o ano: ")
	fmt.Scan(&ano)

	// Solicita ao usuário o número de meses retroativo
	var numIteracoes int
	fmt.Print("Digite o número de meses retroativo: ")
	fmt.Scan(&numIteracoes)

	// Gerar os arquivos a partir da data fornecida
	var pdfs []string
	for i := 0; i < numIteracoes; i++ {
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
		err = os.Mkdir(usuario, 0755) // Cria a pasta com permissões 0755 (rwxr-xr-x)
		if err != nil && !os.IsExist(err) {
			panic(err)
		}

		err = ioutil.WriteFile(caminhoArquivo, content, 0644)
		if err != nil {
			panic(err)
		}
		pdfs = append(pdfs, caminhoArquivo)

		// Decrementa o mês e o ano
		if mes == 1 {
			mes = 12
			ano--
		} else {
			mes--
		}
	}

	// Caminho completo para o executável do PDFtk
	pdftkPath := "C:\\Program Files (x86)\\PDFtk\\bin\\pdftk.exe"

	// Define o nome do arquivo mesclado usando o nome de login
	nomeArquivoMesclado := fmt.Sprintf("%s.pdf", usuario)

	// Define o caminho completo para salvar o arquivo mesclado dentro da pasta do usuário
	caminhoArquivoMesclado := fmt.Sprintf("%s/%s", usuario, nomeArquivoMesclado)

	// Realiza a mesclagem dos PDFs usando o PDFtk
	err := mergePDFs(pdftkPath, pdfs, caminhoArquivoMesclado)
	if err != nil {
		panic(err)
	}

	fmt.Println("Arquivos PDF mesclados com sucesso.")

	// Apaga os arquivos unitários
	for _, arquivo := range pdfs {
		err := os.Remove(arquivo)
		if err != nil {
			fmt.Printf("Erro ao excluir o arquivo %s: %s\n", arquivo, err)
		}
	}

	fmt.Println("Arquivos unitários excluídos com sucesso.")
}

// Função para mesclar os arquivos PDF usando o PDFtk
func mergePDFs(pdftkPath string, pdfs []string, outputFile string) error {
	// Monta o comando para mesclar os PDFs
	args := []string{}
	args = append(args, pdfs...)
	args = append(args, "cat", "output", outputFile)

	// Executa o comando no CMD usando o pdftk
	cmd := exec.Command("cmd.exe", "/c", pdftkPath)
	cmd.Args = append(cmd.Args, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
