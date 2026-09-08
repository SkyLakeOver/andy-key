package frontend

// IndexHTML объединяет все части фронтенда в правильном порядке
// ПОРЯДОК КРИТИЧЕН: сначала HTML, затем все JS части, в конце закрывающие теги
var IndexHTML = uiHTML + initJS + referencesJS + tasksJS + credentialsJS + scriptsJS + bashJS + logsJS + `</script></body></html>`