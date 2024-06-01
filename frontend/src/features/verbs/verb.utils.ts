import { JsonValue } from 'type-fest/source/basic'
import { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { MetadataCalls, MetadataCronJob, MetadataIngress, Ref } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'
import { JSONSchemaFaker } from 'json-schema-faker'

const basePath = 'http://localhost:8891/'

export const verbRefString = (verb: Ref): string => {
  return `${verb.module}.${verb.name}`
}

interface JsonMap { [key: string]: JsonValue }

const processJsonValue = (value: JsonValue): JsonValue =>{
  if (Array.isArray(value)) {
    return value.map(item => processJsonValue(item))
  } else if (typeof value === 'object' && value !== null) {
    const result: JsonMap = {}
    Object.entries(value).forEach(([key, val]) => {
      result[key] = processJsonValue(val)
    })
    return result
  } else if (typeof value === 'string') {
    return ''
  } else if (typeof value === 'number') {
    return 0
  } else if (typeof value === 'boolean') {
    return false
  }
  return value
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export const simpleJsonSchema = (verb: Verb): any => {
  let schema = JSON.parse(verb.jsonRequestSchema)

  if (schema.properties && isHttpIngress(verb)) {
    schema = {
      ...schema,
      type: schema.properties.body.type,
      properties: {
        body: schema.properties.body
      },
      required: schema.required.includes('body') ? ['body'] : []
    }
  }

  return schema
}

export const defaultRequest = (verb?: Verb): string => {
  if (!verb || !verb.jsonRequestSchema) {
    return '{}'
  }

  const schema = simpleJsonSchema(verb)

  JSONSchemaFaker.option({
    alwaysFakeOptionals: false,
    useDefaultValue: true,
    requiredOnly: true
  })

  let fake = JSONSchemaFaker.generate(schema)
  if (fake) {
    fake = processJsonValue(fake)
  }

  return JSON.stringify(fake, null, 2) ?? '{}'
}

export const ingress = (verb?: Verb) => {
  return verb?.verb?.metadata?.find(meta => meta.value.case === 'ingress')?.value?.value as MetadataIngress || null
}

export const cron = (verb?: Verb) => {
  return verb?.verb?.metadata?.find(meta => meta.value.case === 'cronJob')?.value?.value as MetadataCronJob || null
}

export const requestType = (verb?: Verb) => {
  const ingress = verb?.verb?.metadata?.find(meta => meta.value.case === 'ingress')?.value?.value as MetadataIngress || null
  const cron = verb?.verb?.metadata?.find(meta => meta.value.case === 'cronJob')?.value?.value as MetadataCronJob || null

  return ingress?.method ?? cron?.cron?.toUpperCase() ?? 'CALL'
}

export const httpRequestPath = (verb?: Verb) => {
  const ingress = verb?.verb?.metadata?.find(meta => meta.value.case === 'ingress')?.value?.value as MetadataIngress || null
  if (ingress) {
    return ingress.path.map(p => {
      switch (p.value.case) {
        case 'ingressPathLiteral':
          return p.value.value.text
        case 'ingressPathParameter':
          return `{${p.value.value.name}}`
        default:
          return ''
      }
    }).join('/')
  }
}

export const fullRequestPath = (module?: Module, verb?: Verb) => {
  const ingress = verb?.verb?.metadata?.find(meta => meta.value.case === 'ingress')?.value?.value as MetadataIngress || null
  if (ingress) {
    return basePath + httpRequestPath(verb)
  }
  const cron = verb?.verb?.metadata?.find(meta => meta.value.case === 'cronJob')?.value?.value as MetadataCronJob || null
  if (cron) {
    return cron.cron
  }

  return [module?.name, verb?.verb?.name].filter(Boolean).map(v => v).join('.')
}

export const httpPopulatedRequestPath = ( module?: Module, verb?: Verb) => {
  return fullRequestPath(module, verb).replaceAll(/{([^}]*)}/g, '$1')
}

export const isHttpIngress = (verb?: Verb) => {
  return ingress(verb)?.type === 'http'
}

export const isCron = (verb?: Verb) => {
  return !!cron(verb)
}

export const isExported = (verb?: Verb) => {
  return verb?.verb?.export === true
}

export const createVerbRequest = (path: string, verb?: Verb,  editorText?: string, headers?: string) => {
  if (!verb || !editorText) {
    return new Uint8Array()
  }

  const requestJson = JSON.parse(editorText)

  if (isHttpIngress(verb)) {
    const httpIngress = ingress(verb)
    if (httpIngress) {
      requestJson['method'] = httpIngress.method
      requestJson['path'] = path.replace(basePath, '')
    }
    requestJson.headers = JSON.parse(headers ?? '{}')
  }

  const buffer = Buffer.from(JSON.stringify(requestJson))
  return new Uint8Array(buffer)
}

export const verbCalls = (verb?: Verb) => {
  return verb?.verb?.metadata
    .filter((meta) => meta.value.case === 'calls')
    .map((meta) => meta.value.value as MetadataCalls) ?? null
}
