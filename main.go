package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pdfcpu/pdfcpu/pkg/api"
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
		startMonth, _ := strconv.Atoi(entryMes.Text)
		startYear, _ := strconv.Atoi(entryAno.Text)
		numIteracoes, _ := strconv.Atoi(entryNumIteracoes.Text)

		date := time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)

		// Novo navegador
		browser := rod.New().NoDefaultDevice().MustConnect()

		// Fecha o navegador quando a rotina finalizar (defer)
		defer browser.Close()

		// Inicializar o navegador e abrir a página de login
		page := browser.MustPage("https://www.fazenda.sp.gov.br/folha/nova_folha/acessar_dce.asp?menu=dem&user=rs").MustWaitLoad()

		// Preencher os campos de login
		page.MustElement("input[name='txt_logindce']").MustInput(usuario)
		page.MustElement("input[name='txt_senhadce']").MustInput(senha)

		// Clicar no botão de login
		page.MustElement("input[type='submit'][name='enviar']").MustClick()

		// Aguardar a página carregar
		page.MustWaitLoad()

		// Clicar no botão OK do popup de abertura
		page.MustElement("input[type='button'][value='OK']").MustClick()

		// Gerar os arquivos a partir da data fornecida
		var pdfs []string

		// flag para indicar se possui mais páginas
		has_more_pages := false

		// flag para indicar que já está na segunda página
		second_page := false

		// contador de interações desde a última mudança de ano
		interactions_for_second_page_reset := 0

		for i := 0; i < numIteracoes; i++ {
			dataAtual := date.Format("01/2006")

			fmt.Printf("Start %s\n", dataAtual)

			var url string

			if !second_page {
				url = fmt.Sprintf("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_listar.asp?opcao=listar&acao=visualizar&ano=%v", date.Year())
			} else {
				url = fmt.Sprintf("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_listar.asp?opcao=listar&acao=visualizar&page=2&ano=%v", date.Year())
			}

			fmt.Printf("Loading page: %s\n", url)

			// Navegar para a página correspondente
			page.Navigate(url)

			// Aguardar o carregamento da página
			page.MustWaitLoad()

			// flag utilizada indicar se essa iteração deve ser ignorada (Exemplo: mês sem dado ou somente com décimo terceiro)
			ignorar_interacao := true

			// carrega todos o html da página
			html_content, _ := page.HTML()

			// verificar se existe algum link com indicando a segunda página
			if strings.Contains(html_content, "&amp;page=2&amp;") {
				// flag para indicar se existe uma segunda página
				has_more_pages = true
			}

			// Captura todos os links da página
			for _, link := range page.MustElements("a") {
				txt, err := link.Text()
				if err != nil {
					panic(fmt.Sprint("Erro ao obter o texto do elemento:", err))
				}

				if !strings.Contains(txt, dataAtual) {
					continue
				}

				href, err := link.Attribute("href")
				if err != nil {
					panic(fmt.Sprint("Erro ao obter o atributo href do elemento:", err))
				}

				// É décimo terceiro, ignora
				if strings.Contains(*href, "&tp=8&") || strings.Contains(*href, "&tp=5&") {
					continue
				}

				fmt.Println("Link encontrado:", *href)

				new_link := strings.Split(*href, "?")[1]
				new_link = "https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_imp.asp?" + new_link + "&modo=imprimir"

				fmt.Println("Novo link:", new_link)

				// Navegar para a página do link e espera carregar
				new_page := browser.MustPage(new_link).MustWaitLoad()

				// Cria o pdf a partir da página
				pdf, err := new_page.PDF(&proto.PagePrintToPDF{})
				if err != nil {
					panic(fmt.Sprint("Erro ao salvar o pdf:", err))
				}

				// Lê o conteúdo do PDF
				content, err := io.ReadAll(pdf)
				if err != nil {
					panic(fmt.Sprint("Erro ao ler conteúdo do pdf:", err))
				}

				pdf_filename := fmt.Sprintf("./%s.pdf", date.Format("2006-01"))

				// Cria o arquivo
				new_file, err := os.Create(pdf_filename)
				if err != nil {
					panic(fmt.Sprint("Erro ao criar o arquivo:", err))
				}

				// Escreve os dados da página no pdf
				_, err = new_file.Write(content)
				if err != nil {
					panic(fmt.Sprint("Erro ao salvar o pdf:", err))
				}

				// Sincroniza os dados com o diso e fecha o arquivo
				new_file.Sync()
				new_file.Close()

				pdfs = append(pdfs, pdf_filename)

				ignorar_interacao = false
			}

			if ignorar_interacao {
				i--
			}

			oldYear := date.Year()
			interactions_for_second_page_reset++

			date = date.AddDate(0, -1, 0)

			// Se houve mudança de ano e houver mais páginas, volta um ano e ativa a flag de segunda página para mudar a url principal
			if date.Year() != oldYear {
				if has_more_pages && !second_page {
					second_page = true
					date = date.AddDate(0, interactions_for_second_page_reset, 0)
				} else if second_page {
					second_page = false
				}

				interactions_for_second_page_reset = 0
			}

			has_more_pages = false

			fmt.Println()
		}

		// ordena os pdfs em ordem crescente
		sort.Strings(pdfs)

		merge_all_pdfs(usuario, pdfs)

		delete_all_files(pdfs)
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

func merge_all_pdfs(filename string, files []string) {
	err := api.MergeCreateFile(files, fmt.Sprintf("%s.pdf", filename), nil)
	if err != nil {
		panic(fmt.Sprint("Erro ao juntar os pdfs:", err))
	}
}

func delete_all_files(files []string) {
	for _, arquivo := range files {
		err := os.Remove(arquivo)
		if err != nil {
			fmt.Printf("Erro ao excluir um arquivo: %s\n", err)
		}
	}
}
