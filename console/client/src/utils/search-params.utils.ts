export const urlSearchParamsToObject = (params: URLSearchParams) =>  {
  return [ ...params.entries() ].reduce((acc, [ key, value ]) => {
    acc[key] = value
    return acc
  }, {})
}
