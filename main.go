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
		pdfs := []string{}

		// todos os links encontrados
		all_links := []string{}

		// para controlar se houve troca de ano, se houver carrega todos os links do novo ano
		load_all_links := true

		// Na primeira fase, ele vai interar para pegar todos os links de todos os anos dentro do range das iterações
		for i := 0; i < numIteracoes; i++ {
			dataAtual := date.Format("01/2006")

			fmt.Printf("Start %s\n", dataAtual)

			if load_all_links {
				fmt.Println("Loading all links")

				// carrega todos os links da página 1 desse ano
				all_links = append(all_links, all_links_from_year_and_page(page, date.Year(), 1)...)

				// carrega todos os links da página 2 desse ano (se houver, senão ele vai recarregar a página 1, vamos incluir todos os links mesmo assim, ao final iremos higienizar os duplicados)
				all_links = append(all_links, all_links_from_year_and_page(page, date.Year(), 2)...)

				load_all_links = false
			}

			oldYear := date.Year()

			date = date.AddDate(0, -1, 0)

			if date.Year() != oldYear {
				load_all_links = true
			}

			fmt.Println()
		}

		// Terminada a primeira fase, vamos higienizar possíveis links duplicados e ordenar
		new_links := []string{}
		for _, link := range all_links {
			found := false

			for _, new_link := range new_links {
				if link == new_link {
					found = true
					break
				}
			}

			if !found {
				new_links = append(new_links, link)
			}
		}

		sort.Strings(new_links)

		// Reinicia a data e refaz o processo de iterar, porém agora usando os links já prontos
		date = time.Date(startYear, time.Month(startMonth), 1, 0, 0, 0, 0, time.UTC)

		// Nessa fase, ele vai interar para pegar todos os links de todos os anos dentro do range das iterações
		for i := 0; i < numIteracoes; i++ {
			dataAtual := date.Format("01/2006")

			fmt.Printf("Start %s\n", dataAtual)

			// para caso tenha mais de um arquivo dentro do mesmo mês, um não substituir o outro
			count := 0

			// procura os links com a data atual
			for _, link := range new_links {
				if strings.Contains(link, date.Format("0601")) {
					count++

					fmt.Println("Found link:", link)

					// Navegar para a página do link e espera carregar
					new_page := browser.MustPage(link).MustWaitLoad()

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

					pdf_filename := fmt.Sprintf("./%s-%v.pdf", date.Format("2006-01"), count)

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
				}
			}

			date = date.AddDate(0, -1, 0)
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

func all_links_from_year_and_page(page *rod.Page, year, pagination int) []string {
	result := []string{}

	url := fmt.Sprintf("https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_listar.asp?opcao=listar&acao=visualizar&ano=%v&page=%v", year, pagination)

	// Navegar para a página correspondente
	page.Navigate(url)

	// Aguardar o carregamento da página
	page.MustWaitLoad()

	// Captura todos os links da página
	for _, link := range page.MustElements("a") {
		href, err := link.Attribute("href")
		if err != nil {
			panic(fmt.Sprint("Erro ao obter o atributo href do elemento:", err))
		}

		// Se é folha normal, registra
		if strings.Contains(*href, "&tp=0&") {
			// já prepara o link no novo formato
			result = append(result, "https://www.fazenda.sp.gov.br/folha/nova_folha/dem_pagto_imp.asp?"+strings.Split(*href, "?")[1]+"&modo=imprimir")
		}
	}

	return result
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
