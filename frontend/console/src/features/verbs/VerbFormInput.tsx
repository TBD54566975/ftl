export const VerbFormInput = ({
  requestType,
  path,
  setPath,
  requestPath,
  readOnly,
  onSubmit,
}: {
  requestType: string
  path: string
  setPath: (path: string) => void
  requestPath: string
  readOnly: boolean
  onSubmit: (path: string) => void
}) => {
  const handleSubmit: React.FormEventHandler<HTMLFormElement> = async (event) => {
    event.preventDefault()
    onSubmit(path)
  }

  return (
    <form onSubmit={handleSubmit} className='rounded-lg'>
      <div className='flex rounded-md shadow-sm'>
        <span id='call-type' className='inline-flex items-center rounded-l-md border border-r-0 border-gray-300 dark:border-gray-500 px-3 ml-4 sm:text-sm'>
          {requestType}
        </span>
        <input
          type='text'
          name='request-path'
          id='request-path'
          className='block w-full min-w-0 flex-1 rounded-none rounded-r-md border-0 py-1.5 dark:bg-transparent ring-1 ring-inset ring-gray-300 dark:ring-gray-500 focus:ring-2 focus:ring-inset focus:ring-indigo-600 sm:text-sm sm:leading-6'
          value={path}
          readOnly={readOnly}
          onChange={(event) => setPath(event.target.value)}
        />
        <button type='submit' className='bg-indigo-700 text-white ml-2 mr-4 px-4 py-2 rounded-lg hover:bg-indigo-600 focus:outline-none focus:bg-indigo-600'>
          Send
        </button>
      </div>
      {!readOnly && <span className='ml-4 text-xs text-gray-500'>{requestPath}</span>}
    </form>
  )
}
