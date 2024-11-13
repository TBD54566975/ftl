import { useEffect, useState } from 'react'
import { useParams } from 'react-router-dom'
import { formatMoney } from '../../utils/money.utils'
import { Product, ProductcatalogClient } from '../../api/productcatalog'

export const ProductPage = () => {
  const { productId } = useParams()
  const [product, setProduct] = useState<Product | undefined>()

  useEffect(() => {
    const productsClient = new ProductcatalogClient('http://localhost:8891')
    productsClient.list({}).then((response) => {
      setProduct(response.products.find((product) => product.id.toLowerCase() === productId?.toLowerCase()))
    })
  }, [])

  return (
    <div className='bg-white'>
      <div className='pb-16 pt-6 sm:pb-24'>
        <div className='mx-auto mt-8 max-w-2xl px-4 sm:px-6 lg:max-w-7xl lg:px-8'>
          <div className='lg:grid lg:auto-rows-min lg:grid-cols-12 lg:gap-x-8'>
            <div className='lg:col-span-5 lg:col-start-8'>
              <div className='flex justify-between'>
                <h1 className='text-xl font-medium text-gray-900'>{product?.name}</h1>
                {product?.priceUsd && (
                  <p className='text-xl font-medium text-gray-900'>{formatMoney(product.priceUsd)}</p>
                )}
              </div>
              <div className='mt-4'>
                {product?.categories.map((category) => (
                  <span
                    key={category}
                    className='inline-flex items-center rounded-md bg-gray-100 px-2 py-1 text-xs font-medium text-gray-600 mr-2'
                  >
                    {category}
                  </span>
                ))}
              </div>

              <div className='mt-10'>
                <h2 className='text-sm font-medium text-gray-900'>Description</h2>

                <div
                  className='prose prose-sm mt-4 text-gray-500'
                  dangerouslySetInnerHTML={{ __html: product?.description ?? '' }}
                />
              </div>

              <div className='mt-8 lg:col-span-5'>
                <form>
                  <button
                    type='submit'
                    className='mt-8 flex w-full items-center justify-center rounded-md border border-transparent bg-indigo-600 px-8 py-3 text-base font-medium text-white hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2'
                  >
                    Add to cart
                  </button>
                </form>
              </div>
            </div>

            <div className='mt-8 lg:col-span-7 lg:col-start-1 lg:row-span-3 lg:row-start-1 lg:mt-0'>
              <div className='grid grid-cols-1 lg:grid-cols-2 lg:grid-rows-3 lg:gap-8'>
                <img
                  key={product?.id}
                  src={product?.picture}
                  alt={product?.name}
                  className='lg:col-span-2 lg:row-span-2 rounded-lg'
                />
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
