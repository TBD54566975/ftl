// Return the corresponding Dart type for an FTL type.
function dartType(t) {
  const type = typename(t);
  switch (type) {
    case "String":
      return "String";

    case "Int":
      return "int";

    case "Bool":
      return "bool";

    case "Float":
      return "double";

    case "Time":
      return "DateTime";

    case "Map":
      return `Map<${dartType(t.key)}, ${dartType(t.value)}>`;

    case "Array":
      return `List<${dartType(t.element)}>`;

    case "Bytes":
      return `Uint8List`;

    case "Ref":
    case "VerbRef":
    case "DataRef":
      if (context.name === t.module) {
        return t.name;
      }
      if (t.typeParameters && t.typeParameters.length > 0) {
        return `${t.module}.${t.name}${dartTypeParameters(t.typeParameters)}`
      }
      return [t.module, t.name].filter(Boolean).join('.');

    case "TypeParameter":
      return t.name;

    case "Optional":
      return dartType(t.type) + "?";

    default:
      throw new Error(`Unspported FTL type: ${typename(t)}`);
  }
}

function dartTypeParameters(t) {
  if (t.length == 0) {
    return "";
  }

  return `<${t.map((p) => dartType(p)).join(", ")}>`;
}

function bodyType(t) {
  return dartType(t.typeParameters[0]);
}

function tpMappingFuncs(tps) {
  if (tps.length === 0) {
    return "";
  }

  let response = "";
  for (let tp of tps) {
    response += `, ${tp} Function(Map<String, dynamic>) ${tp.name.toLowerCase()}JsonFn`;
  }

  return response;
}

// This function returns the fields of a type as a string.
// with special handling for type parameters.
function fromJsonFields(t) {
  let response = "";

  for (let field of t.fields) {
    let isTypeParameter = false;
    for (let tp of t.typeParameters) {
      if (field.type.name === tp.name) {
        response += `${field.name}: ${tp.name.toLowerCase()}JsonFn(map['${field.name}']), `;
        isTypeParameter = true;
        break;

      }
    }
    if (!isTypeParameter) {
      response += `${field.name}: ((dynamic v) => ${deserialize(field.type)})(map['${field.name}']), `;
    }
  }

  return response;
}

function deserialize(t) {
  switch (typename(t)) {
    case "Array":
      return `v.map((v) => ${deserialize(t.element)}).cast<${dartType(t.element)}>().toList()`;

    case "Map":
      return `v.map((k, v) => MapEntry(k, ${deserialize(t.value)})).cast<${dartType(t.key)}, ${dartType(t.value)}>()`;

    case "Ref":
    case "DataRef":
      return `${dartType(t)}.fromJson(v)`;

    default:
      return "v";
  }
}

function serialize(t) {
  switch (typename(t)) {
    case "Array":
      return `v.map((v) => ${serialize(t.element)}).cast<${dartType(t.element)}>().toList()`;

    case "Map":
      return `v.map((k, v) => MapEntry(k, ${serialize(t.value)})).cast<${dartType(t.key)}, ${dartType(t.value)}>()`;

    case "Ref":
    case "DataRef":
      return "v.toJson()";

    default:
      return "v";
  }
}

function url(verb) {
  let path = '/' + verb.metadata[0].path.join("/");

  return path.replace(/{(.*?)}/g, (match, fieldName) => {
    return "$" + `{request.${fieldName}}`;
  });
}
