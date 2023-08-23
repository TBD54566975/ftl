import { useState } from 'react'
import { CodeBlock } from '../../components/CodeBlock'
import { useClient } from '../../hooks/use-client'
import { Module, Verb } from '../../protos/xyz/block/ftl/v1/console/console_pb'
import { VerbService } from '../../protos/xyz/block/ftl/v1/ftl_connect'
import { VerbRef } from '../../protos/xyz/block/ftl/v1/schema/schema_pb'

type Props = {
  module?: Module
  verb?: Verb
}

export const VerbForm: React.FC<Props> = ({ module, verb }) => {
  const client = useClient(VerbService)
  const [ response, setResponse ] = useState<string | null>(null)
  const [ error, setError ] = useState<string | null>(null)

  const callData = module?.data.filter(data =>
    [ verb?.verb?.request?.name, verb?.verb?.response?.name ].includes(data.name)
  )

  const handleSubmit = async event => {
    event.preventDefault()

    setResponse(null)
    setError(null)

    const formData = new FormData(event.target)
    // Convert the form data to a plain object (or however you want to send it)
    const dataObject = Array.from(formData.entries()).reduce((obj, [ key, value ]) => {
      obj[key] = value
      return obj
    }, {})

    try {
      const verbRef: VerbRef = {
        name: verb?.verb?.name,
        module: module?.name,
      } as VerbRef

      const buffer = Buffer.from(JSON.stringify(dataObject))
      const uint8Array = new Uint8Array(buffer)

      try {
        const response = await client.call({ verb: verbRef, body: uint8Array })
        if (response.response.case === 'body') {
          const jsonString = Buffer.from(response.response.value).toString('utf-8')

          setResponse(JSON.stringify(JSON.parse(jsonString), null, 2))
        } else if (response.response.case === 'error') {
          setError(response.response.value.message)
        }
      } catch (error) {
        console.error('There was an error with the request:', error)
      }
    } catch (error) {
      console.error('There was an error with the request:', error)
    }
  }

  return (
    <>
      <form onSubmit={handleSubmit} className='rounded-lg'>
        {callData?.filter(d => d.name === verb?.verb?.request?.name).map((data, dataIndex) => (
          <div key={dataIndex}
            className='mb-4'
          >
            <h2 className='text-lg font-semibold mb-2'>{data.name}</h2>
            {data.fields.map((field, fieldIndex) => (
              <div key={fieldIndex}
                className='text-sm mb-3'
              >
                <label htmlFor={`input-${dataIndex}-${fieldIndex}`}
                  className='block text-sm font-medium mb-1'
                >
                  {field.name}:
                </label>
                <input
                  id={`input-${dataIndex}-${fieldIndex}`}
                  name={field.name}
                  type='text'
                  placeholder={`Enter ${field.name}`}
                  className='w-full text-gray-900 px-3 py-2 border rounded shadow-sm focus:outline-none focus:border-blue-500'
                />
              </div>
            ))}
          </div>
        ))}
        <button type='submit'
          className='bg-indigo-700 text-white px-4 py-2 rounded hover:bg-indigo-600 focus:outline-none focus:bg-indigo-600'
        >
          Submit
        </button>
      </form>
      {response && (
        <CodeBlock code={response} language='go' />
      )}
      {error && (
        <div className='mt-4 bg-red-100 border-l-4 border-red-500 text-red-700 p-4' role='alert'>
          {error}
        </div>
      )}
    </>
  )
}
