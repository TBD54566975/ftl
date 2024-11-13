// Return the corresponding Typescript type for an FTL type.
function tsType(t) {
  const type = typename(t)
  switch (type) {
    case 'String':
      return 'string'

    case 'Int':
      return 'number'

    case 'Bool':
      return 'bool'

    case 'Float':
      return 'double'

    case 'Time':
      return 'DateTime'

    case 'Map':
      return `Map<${tsType(t.key)}, ${tsType(t.value)}>`

    case 'Array':
      return `${tsType(t.element)}[]`

    case "Bytes":
      return "Uint8Array";

    case 'Ref':
    case 'VerbRef':
    case 'DataRef':
      if (context.name === t.module) {
        return t.name
      }
      if (t.typeParameters && t.typeParameters.length > 0) {
        return `${t.module}.${t.name}${tsTypeParameters(t.typeParameters)}`
      }
      return [t.module, t.name].filter(Boolean).join('.');

    case 'Optional':
      return tsType(t.type) + '?'

    case "TypeParameter":
      return t.name;

    default:
      throw new Error(`Unspported FTL type: ${typename(t)}`)
  }
}

function bodyType(t) {
  return tsType(t.typeParameters[0]);
}

function tsTypeParameters(t) {
  if (t.length == 0) {
    return "";
  }

  return `<${t.map((p) => tsType(p)).join(", ")}>`;
}

function deserialize(t) {
  switch (typename(t)) {
    case 'Array':
      return `v.map((v) => ${deserialize(t.element)}).cast<${tsType(t.element)}>().toList()`

    case 'Map':
      return `v.map((k, v) => MapEntry(k, ${deserialize(t.value)})).cast<${tsType(t.key)}, ${tsType(t.value)}>()`

    case 'DataRef':
      return `${tsType(t)}.fromMap(v)`

    default:
      return 'v'
  }
}

function serialize(t) {
  switch (typename(t)) {
    case 'Array':
      return `v.map((v) => ${serialize(t.element)}).cast<${tsType(t.element)}>().toList()`

    case 'Map':
      return `v.map((k, v) => MapEntry(k, ${serialize(t.value)})).cast<${tsType(t.key)}, ${tsType(t.value)}>()`

    case 'DataRef':
      return 'v.toMap()'

    default:
      return 'v'
  }
}

function url(verb) {
  let path = "/" + verb.metadata[0].path.join("/");
  const method = verb.metadata[0].method

  path = path.replace(/{(.*?)}/g, (match, fieldName) => {
    return '$' + `{request.${fieldName}}`
  })

  return method !== 'GET' ? path : path + '?@json=${encodeURIComponent(JSON.stringify(request))}'
}
