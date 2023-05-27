package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

func main() {
	a := app.New()
	window := a.NewWindow("Folha de Pagamento")

	// Campos de entrada
	entryUsuario := widget.NewEntry()
	entrySenha := widget.NewPasswordEntry()
	entryMes := widget.NewEntry()
	entryAno := widget.NewEntry()
	entryNumIteracoes := widget.NewEntry()

	// Botão de execução
	button := widget.NewButton("Executar", func() {
		// Obter os valores dos campos de entrada
		usuario := entryUsuario.Text
		senha := entrySenha.Text
		mes := entryMes.Text
		ano := entryAno.Text
		numIteracoesStr := entryNumIteracoes.Text

		// Converter o número de iterações para o tipo int
		numIteracoes, err := strconv.Atoi(numIteracoesStr)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		// Inicializar o navegador e abrir a página de login
		page := rod.New().NoDefaultDevice().MustConnect().MustPage("https://www.fazenda.sp.gov.br/folha/nova_folha/acessar_dce.asp?menu=dem&user=rs").MustWaitLoad()

		// Preencher os campos de login
		page.MustElement("input[name='txt_logindce']").MustInput(usuario)
		page.MustElement("input[name='txt_senhadce']").MustInput(senha)

		// Clicar no botão de login
		page.MustElement("input[type='submit'][name='enviar']").MustClick()

		// Aguardar a página carregar
		page.MustWaitLoad()

		// Clicar no botão OK do popup de abertura
		page.MustElement("input[type='button'][value='OK']").MustClick()

		// Converter o mês e o ano para o tipo int
		mesInt, err := strconv.Atoi(mes)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}
		anoInt, err := strconv.Atoi(ano)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		// Gerar os arquivos a partir da data fornecida
		var pdfs []string
		for i := 0; i < numIteracoes; i++ {
			// Construir a URL com base no mês, ano e login do usuário
			url := fmt.Sprintf("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_imp.asp?sq=1&tp=0&dt=1%02d%02d&rb=0&rs=%s&nro=0&tabela=atual&sit=1&dt_sit=&pv=01&opcao_pagto=visualizar&tipo_usuario=rs&opcao=abertura&acao=&ver_aviso=true&modo=imprimir", anoInt, mesInt, usuario)
			// Navegar para a página correspondente
			page.Navigate(url)

			// Aguardar 2 segundos
			time.Sleep(2 * time.Second)

			// Salvar a página como PDF
			pdf, err := page.PDF(&proto.PagePrintToPDF{})
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			// Ler o conteúdo do PDF
			content, err := ioutil.ReadAll(pdf)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}

			// Formatar o nome do arquivo
			nomeArquivo := fmt.Sprintf("%02d-%02d.pdf", anoInt, mesInt)

			// Definir o caminho completo para salvar o arquivo dentro da pasta do usuário
			caminhoArquivo := fmt.Sprintf("%s/%s", usuario, nomeArquivo)

			// Escrever o conteúdo do PDF em um arquivo com o novo nome e caminho
			err = os.Mkdir(usuario, 0755) // Criar a pasta com permissões 0755 (rwxr-xr-x)
			if err != nil && !os.IsExist(err) {
				dialog.ShowError(err, window)
				return
			}

			err = ioutil.WriteFile(caminhoArquivo, content, 0644)
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			pdfs = append(pdfs, caminhoArquivo)

			// Decrementar o mês e o ano
			if mesInt == 1 {
				mesInt = 12
				anoInt--
			} else {
				mesInt--
			}
		}

		// Caminho completo para o executável do PDFtk
		pdftkPath := "C:\\Program Files (x86)\\PDFtk\\bin\\pdftk.exe"

		// Definir o nome do arquivo mesclado usando o nome de login
		nomeArquivoMesclado := fmt.Sprintf("%s.pdf", usuario)

		// Definir o caminho completo para salvar o arquivo mesclado dentro da pasta do usuário
		caminhoArquivoMesclado := fmt.Sprintf("%s/%s", usuario, nomeArquivoMesclado)

		// Realizar a mesclagem dos PDFs usando o PDFtk
		err = mergePDFs(pdftkPath, pdfs, caminhoArquivoMesclado)
		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		dialog.ShowInformation("Sucesso", "Arquivos PDF mesclados com sucesso.", window)

		// Excluir os arquivos unitários
		for _, arquivo := range pdfs {
			err := os.Remove(arquivo)
			if err != nil {
				fmt.Printf("Erro ao excluir o arquivo %s: %s\n", arquivo, err)
			}
		}

		dialog.ShowInformation("Sucesso", "Arquivos unitários excluídos com sucesso.", window)
	})

	// Layout dos campos de entrada
	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Usuário:", Widget: entryUsuario},
			{Text: "Senha:", Widget: entrySenha},
			{Text: "Mês:", Widget: entryMes},
			{Text: "Ano:", Widget: entryAno},
			{Text: "Número de meses retroativo:", Widget: entryNumIteracoes},
		},
		OnSubmit: func() {
			button.OnTapped()
		},
	}

	// Layout principal
	content := fyne.NewContainerWithLayout(
		layout.NewVBoxLayout(),
		form,
		button,
	)

	window.SetContent(content)
	window.ShowAndRun()
}

// Função para mesclar os arquivos PDF usando o PDFtk
func mergePDFs(pdftkPath string, pdfs []string, outputFile string) error {
	// Montar o comando para mesclar os PDFs
	args := []string{}
	args = append(args, pdfs...)
	args = append(args, "cat", "output", outputFile)

	// Executar o comando no CMD usando o pdftk
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
