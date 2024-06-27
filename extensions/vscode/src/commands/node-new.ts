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

  const nodeType = await vscode.window.showQuickPick(['verb', 'ingress:http', 'enum', 'pubsub', 'pubsub:subscription', 'fsm', 'database', 'config:string', 'config:struct', 'secret', 'cron'], {
    title: 'Which type of node would you like to add',
    placeHolder: 'Choose a node type',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (nodeType === undefined) {
    return
  }

  const snippet = await snippetForNodeType(nodeType, item)

  if (snippet === '') {
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
      return enumSnippet

    case 'pubsub':
      return publisherSnippet + '\n\n' + subscriberSnippet

    case 'pubsub:subscription':
      return subscriberSnippet

    case 'fsm':
      return fsmSnippet

    case 'database':
      return databaseSnippet()

    case 'config:string':
      return `var sampleConfig = ftl.Config[string]("sample_config")`

    case 'config:struct':
      return configStructSnippet

    case 'secret':
      return `var sampleSecret = ftl.Secret[string]("sample_secret")`

    case 'cron':
      return cronSnippet
  }

  return ''
}

const verbSnippet = async () => {
  const inputBoxOptions: vscode.InputBoxOptions = {
    prompt: 'What would you like to name the verb?',
    placeHolder: 'MyVerb',
  }
  const name = await vscode.window.showInputBox(inputBoxOptions)
  if (name === undefined) {
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
  const inputBoxOptions: vscode.InputBoxOptions = {
    prompt: 'What would you like to name the ingress?',
    placeHolder: 'MyEndpoint',
  }
  const name = await vscode.window.showInputBox(inputBoxOptions)
  if (name === undefined) {
    return undefined
  }

  const method = await vscode.window.showQuickPick(['GET', 'POST', 'PUT', 'DELETE'], {
    title: 'What method will you use for the ingress?',
    placeHolder: 'Choose a method',
    canPickMany: false,
    ignoreFocusOut: true
  })

  if (method === undefined) {
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

const enumSnippet = `//ftl:enum
type SampleEnum string
const (
	FirstValue SampleEnum = "first"
	SecondValue SampleEnum = "second"
)`

const publisherSnippet = `//ftl:export
var sampleTopic = ftl.Topic[SamplePubSubEvent]("sample_topic")

type SamplePubSubEvent struct {
	Message string
}`

const subscriberSnippet = `var _ = ftl.Subscription(sampleTopic, "sample_subscription")

//ftl:verb
//ftl:subscribe sample_subscription
func SampleSubscriber(ctx context.Context, event SamplePubSubEvent) error {
	return nil
}`

const databaseSnippet = async () => {
  const inputBoxOptions: vscode.InputBoxOptions = {
    prompt: 'What would you like to name the database?',
    placeHolder: 'my_db',
  }
  const name = await vscode.window.showInputBox(inputBoxOptions)
  if (name === undefined) {
    return undefined
  }

  return `var ${name.toLowerCase()}Database = ftl.PostgresDatabase("${name.toLowerCase()}")`
}

const fsmSnippet = `type SampleFSMMessage struct {
	Instance string
	Message string
}

var sampleFsm = ftl.FSM("sample_fsm",
	ftl.Start(SampleFSMState0),
	ftl.Transition(SampleFSMState0, SampleFSMState1),
	ftl.Transition(SampleFSMState1, SampleFSMState2),
)

//ftl:verb
func SampleFSMState0(ctx context.Context, in SampleFSMMessage) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("message %q entering state 0", in.Message)
	return nil
}

//ftl:verb
func SampleFSMState1(ctx context.Context, in SampleFSMMessage) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("message %q entering state 1", in.Message)
	return nil
}

//ftl:verb
func SampleFSMState2(ctx context.Context, in SampleFSMMessage) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("message %q entering state 2", in.Message)
	return nil
}

//ftl:verb
func SendSampleFSMMessage(ctx context.Context, in SampleFSMMessage) error {
	return sampleFsm.Send(ctx, in.Instance, in)
}
`

const configStructSnippet = `type SampleConfig struct {
	Field string
}

var sampleConfigValue = ftl.Config[SampleConfig]("sample_config")`

const cronSnippet = `// This cron job will run every 5 minutes
//ftl:cron * /5 * * * * *
func SampleCron(ctx context.Context) error {
	logger := ftl.LoggerFromContext(ctx)
	logger.Infof("sample cron job triggered")
	return nil
}`
