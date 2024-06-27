import * as vscode from 'vscode'
import { FtlTreeItem } from '../tree-item'

export const nodeNewCommand = async (item: FtlTreeItem) => {
  const uri = vscode.Uri.file(item.position?.filename || '')
  let document: vscode.TextDocument

  try {
    document = await vscode.workspace.openTextDocument(uri)
  } catch (error) {
    vscode.window.showErrorMessage(`Error opening file ${item.position?.filename}: ${error}`)
    return
  }

  if (document === undefined) {
    return
  }

  if (document.isDirty) {
    vscode.window.showWarningMessage('Please save the document before adding a new node')
    return
  }

  const nodeType = await vscode.window.showQuickPick(['verb', 'ingress:http', 'enum', 'pubsub', 'fsm', 'database', 'config:value', 'config:struct', 'secret', 'cron'], {
    title: 'Which type of node would you like to add',
    placeHolder: 'Choose a node type',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (!nodeType) {
    return
  }

  const snippet = await snippetForNodeType(nodeType, item)

  if (!snippet) {
    vscode.window.showErrorMessage(`No snippet available for node type ${nodeType}`)
    return
  }

  const editor = await vscode.window.showTextDocument(document)
  const edit = new vscode.WorkspaceEdit()
  const position = new vscode.Position(document.lineCount, 0)
  edit.insert(uri, position, `\n${snippet}`)

  await vscode.workspace.applyEdit(edit)
  await document.save()

  // Scroll to the bottom of the document
  const lastLine = document.lineCount - 1
  const lastLineRange = document.lineAt(lastLine).range
  editor.revealRange(lastLineRange, vscode.TextEditorRevealType.Default)
}

const snippetForNodeType = async (nodeType: string, item: FtlTreeItem): Promise<string | undefined> => {
  switch (nodeType) {
    case 'verb':
      return verbSnippet()

    case 'ingress:http':
      return ingressSnippet(item)

    case 'enum':
      return await enumSnippet()

    case 'pubsub':
      return publisherSnippet()

    case 'fsm':
      return fsmSnippet()

    case 'database':
      return databaseSnippet()

    case 'config:value':
      return configValueSnippet()

    case 'config:struct':
      return configStructSnippet()

    case 'secret':
      return secretSnippet()

    case 'cron':
      return cronSnippet()
  }

  return ''
}

const getTemplateArgument = async (prompt: string, placeHolder: string): Promise<string> => {
  const inputBoxOptions: vscode.InputBoxOptions = {
    prompt: prompt,
    placeHolder: placeHolder,
  }

  return await vscode.window.showInputBox(inputBoxOptions) || ''
}

const snakeToCamel = (snake: string): string => {
  return snake.replace(/(_\w)/g, (m) => m[1].toUpperCase())
}

const snakeToPascal = (snake: string): string => {
  return snake.charAt(0).toUpperCase() + snakeToCamel(snake).slice(1)
}

const verbSnippet = async () => {
  const name = await getTemplateArgument('What would you like to name the verb?', 'MyVerb')
  if (!name) {
    return undefined
  }

  return `type ${name}Request struct {}

type ${name}Response struct {
  Message string
}

//ftl:verb
func ${name}(ctx context.Context, req ${name}Request) (${name}Response, error) {
	return ${name}Response{Message: "Hello, world!"}, nil
}`
}

const ingressSnippet = async (item: FtlTreeItem) => {
  const name = await getTemplateArgument('What would you like to name the ingress?', 'MyEndpoint')
  if (!name) {
    return undefined
  }

  const method = await vscode.window.showQuickPick(['GET', 'POST', 'PUT', 'DELETE'], {
    title: 'What method will you use for the ingress?',
    placeHolder: 'Choose a method',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (!method) {
    return undefined
  }

  return `type ${name}Request struct {}
type ${name}Response struct {}

//ftl:ingress ${method} /${item.moduleName}/${name.toLowerCase()}
func ${name}(ctx context.Context, req builtin.HttpRequest[${name}Request]) (builtin.HttpResponse[${name}Response, ftl.Unit], error) {
	return builtin.HttpResponse[${name}Response, ftl.Unit]{
		Body: ftl.Some(${name}Response{}),
	}, nil
}
`
}

const enumSnippet = async () => {
  const name = await getTemplateArgument('What would you like to name the enum?', 'MyEnum')
  if (!name) {
    return undefined
  }

  return `//ftl:enum
type ${name} string
const (
	FirstValue ${name} = "first"
	SecondValue ${name} = "second"
)`
}

const publisherSnippet = async () => {
  const topic = await getTemplateArgument('What would you like to name the topic?', 'my_topic')
  const event = await getTemplateArgument('What would you like to name the event for this topic?', 'MyEvent')
  if (!topic || !event) {
    return undefined
  }

  const subscription = await getTemplateArgument('What would you like to name the subscription?', `${topic}_subscription`)
  if (!subscription) {
    return undefined
  }
  const subscriber = await getTemplateArgument('What would you like to name the subscriber?', `${snakeToPascal(topic)}Subscriber`)
  if (!subscriber) {
    return undefined
  }

  return `//ftl:export
var ${snakeToCamel(topic)} = ftl.Topic[${event}]("${topic}")

type ${event} struct {
	Message string
}
  
var _ = ftl.Subscription(${snakeToCamel(topic)}, "${subscription}")

//ftl:verb
//ftl:subscribe ${subscription}
func ${subscriber}(ctx context.Context, event ${event}) error {
	return nil
}`
}

const databaseSnippet = async () => {
  const name = await getTemplateArgument('What would you like to name the database?', 'my_db')
  if (!name) {
    return undefined
  }

  return `var ${snakeToCamel(name)}Database = ftl.PostgresDatabase("${name.toLowerCase()}")`
}

const fsmSnippet = async () => {
  const name = await getTemplateArgument('What would you like to name the fsm?', 'my_fsm')
  if (!name) {
    return undefined
  }
  const message = await getTemplateArgument('What would you like to message type for this fsm?', `${snakeToPascal(name)}Message`)
  if (!message) {
    return undefined
  }
  const dispatcher = await getTemplateArgument('What would you like to name the message sender for this fsm?', `Send${snakeToPascal(name)}Message`)

  return `type ${message} struct {
	Instance string
}

var ${snakeToCamel(name)} = ftl.FSM("${name}",
	ftl.Start(${snakeToPascal(name)}State0),
	ftl.Transition(${snakeToPascal(name)}State0, ${snakeToPascal(name)}State1),
	ftl.Transition( ${snakeToPascal(name)}State1,  ${snakeToPascal(name)}State2),
)

//ftl:verb
func ${snakeToPascal(name)}State0(ctx context.Context, msg ${message}) error {
	ftl.LoggerFromContext(ctx).Infof("%q entered state 0", msg.Instance)
	return nil
}

//ftl:verb
func ${snakeToPascal(name)}State1(ctx context.Context, msg ${message}) error {
	ftl.LoggerFromContext(ctx).Infof("%q entered state 1", msg.Instance)
	return nil
}

//ftl:verb
func ${snakeToPascal(name)}State2(ctx context.Context, msg ${message}) error {
	ftl.LoggerFromContext(ctx).Infof("%q entered state 2", msg.Instance)
	return nil
}

//ftl:verb
func ${dispatcher}(ctx context.Context, msg ${message}) error {
	return  ${snakeToCamel(name)}.Send(ctx, msg.Instance, msg)
}
`
}

const configValueSnippet = async () => {
  const name = await getTemplateArgument('What would you like to name the setting for this config?', 'my_config')
  const type = await vscode.window.showQuickPick(['string', 'bool', 'int'], {
    title: 'What value type would you like to assign to this config?',
    placeHolder: 'Choose a value type',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (!name || !type) {
    return undefined
  }

  return `var ${snakeToCamel(name)} = ftl.Config[${type}]("${name}")`
}

const configStructSnippet = async() => {
  const name = await getTemplateArgument('What would you like to name the setting for this config?', 'my_config')
  if (!name) {
    return undefined
  }
  const type = await getTemplateArgument('What would you like to name the struct for this config?', `${snakeToPascal(name)}Config`)
  if (!type) {
    return undefined
  }

  return `type ${type} struct {
	Setting1 string
}

var ${snakeToCamel(name)} = ftl.Config[${type}]("${name}")`
}

const secretSnippet = async () => {
  const name = await getTemplateArgument('What would you like to name the setting for this secret?', 'my_secret')
  const type = await vscode.window.showQuickPick(['string', 'bool', 'int'], {
    title: 'What value type would you like to assign to this secret?',
    placeHolder: 'Choose a value type',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (!name || !type) {
    return undefined
  }

  return `var ${snakeToCamel(name)} = ftl.Secret[${type}]("${name}")`
}

const cronSnippet = async() => {
  const name = await getTemplateArgument('What would you like to name the cron task?', 'MyCronTask')
  const schedule = await getTemplateArgument('What schedule would you like to set for this cron task?', '*/5 * * * *')
  if (!name || !schedule) {
    return undefined
  }
  return `//ftl:cron ${schedule}
func ${name}(ctx context.Context) error {
	ftl.LoggerFromContext(ctx).Infof("sample cron job triggered")
	return nil
}`
}
