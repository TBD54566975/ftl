export default function ModuleNotFound({ id }) {
  return (
    <>
      <div className="px-6 py-24 sm:px-6 sm:py-32 lg:px-8">
        <div className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-gray-900 dark:text-white sm:text-4xl">
            Module not found
            <br />
            <span className="text-indigo-600">{id}</span>
          </h2>
        </div>
      </div>
    </>
  )
}
