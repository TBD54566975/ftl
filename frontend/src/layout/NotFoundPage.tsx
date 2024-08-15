import { ListBulletIcon, Square3Stack3DIcon } from '@heroicons/react/24/outline'

export const NotFoundPage = () => {
  return (
    <div className='bg-white'>
      <main className='mx-auto w-full max-w-7xl px-6 pb-16 pt-10 sm:pb-24 lg:px-8'>
        <div className='mx-auto mt-20 max-w-2xl text-center sm:mt-24'>
          <p className='text-base font-semibold leading-8 text-indigo-600'>404</p>
          <h1 className='mt-4 text-3xl font-bold tracking-tight text-gray-900 sm:text-5xl'>This page does not exist</h1>
          <p className='mt-4 text-base leading-7 text-gray-600 sm:mt-6 sm:text-lg sm:leading-8'>Sorry, we couldn’t find the page you’re looking for.</p>
        </div>
        <div className='mx-auto mt-16 flow-root max-w-lg sm:mt-20'>
          <h2 className='sr-only'>Popular pages</h2>
          <ul className='-mt-6 divide-y divide-gray-900/5 border-b border-gray-900/5'>
            <li className='relative flex gap-x-6 py-6'>
              <div className='flex h-10 w-10 flex-none items-center justify-center rounded-lg shadow-sm ring-1 ring-gray-900/10'>
                <div className='h-6 w-6 text-indigo-600'>
                  <ListBulletIcon />
                </div>
              </div>
              <div className='flex-auto'>
                <h3 className='text-sm font-semibold leading-6 text-gray-900'>
                  <a href='/events'>
                    <span className='absolute inset-0' aria-hidden='true' />
                    Events
                  </a>
                </h3>
                <p className='mt-0 text-sm leading-6 text-gray-600'>View and filter FTL events.</p>
              </div>
              <div className='flex-none self-center'>
                <svg className='h-5 w-5 text-gray-400' viewBox='0 0 20 20' fill='currentColor' aria-hidden='true'>
                  <path
                    fillRule='evenodd'
                    d='M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z'
                    clipRule='evenodd'
                  />
                </svg>
              </div>
            </li>
            <li className='relative flex gap-x-6 py-6'>
              <div className='flex h-10 w-10 flex-none items-center justify-center rounded-lg shadow-sm ring-1 ring-gray-900/10'>
                <div className='h-6 w-6 text-indigo-600'>
                  <Square3Stack3DIcon />
                </div>
              </div>
              <div className='flex-auto'>
                <h3 className='text-sm font-semibold leading-6 text-gray-900'>
                  <a href='/deployments'>
                    <span className='absolute inset-0' aria-hidden='true' />
                    Deployments
                  </a>
                </h3>
                <p className='mt-0 text-sm leading-6 text-gray-600'>View and filter FTL events.</p>
              </div>
              <div className='flex-none self-center'>
                <svg className='h-5 w-5 text-gray-400' viewBox='0 0 20 20' fill='currentColor' aria-hidden='true'>
                  <path
                    fillRule='evenodd'
                    d='M7.21 14.77a.75.75 0 01.02-1.06L11.168 10 7.23 6.29a.75.75 0 111.04-1.08l4.5 4.25a.75.75 0 010 1.08l-4.5 4.25a.75.75 0 01-1.06-.02z'
                    clipRule='evenodd'
                  />
                </svg>
              </div>
            </li>
          </ul>
          <div className='mt-10 flex justify-center'>
            <a href='/' className='text-sm font-semibold leading-6 text-indigo-600'>
              <span aria-hidden='true'>&larr;</span>
              Back to home
            </a>
          </div>
        </div>
      </main>
    </div>
  )
}
