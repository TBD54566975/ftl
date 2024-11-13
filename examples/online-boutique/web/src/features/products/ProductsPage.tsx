import { useEffect, useState } from 'react'
import { Product, ProductcatalogClient } from '../../api/productcatalog'
import { formatMoney } from '../../utils/money.utils'
import { Link } from 'react-router-dom'

export const ProductsPage = () => {
  const [products, setProducts] = useState<Product[]>([])

  useEffect(() => {
    const productsClient = new ProductcatalogClient('http://localhost:8891')
    productsClient.list({}).then((response) => setProducts(response.products || []))
  }, [])

  return (
    <div className='mx-auto max-w-2xl px-4 py-16 sm:px-6 lg:max-w-7xl lg:px-8'>
      <h2 className='text-2xl font-bold tracking-tight text-gray-900'>Products</h2>
      <div className='mt-6 grid grid-cols-1 gap-x-6 gap-y-10 sm:grid-cols-2 lg:grid-cols-4 xl:gap-x-8'>
        {products.map((product) => (
          <Link to={`/products/${product.id.toLowerCase()}`} key={product.id}>
            <div className='group relative'>
              <div className='aspect-h-1 aspect-w-1 w-full overflow-hidden rounded-md bg-gray-200 lg:aspect-none group-hover:opacity-75 lg:h-80'>
                <img
                  src={product.picture}
                  alt={product.description}
                  className='h-full w-full object-cover object-center lg:h-full lg:w-full'
                />
              </div>
              <div className='mt-4 flex justify-between'>
                <div>
                  <h3 className='text-sm text-gray-700'>
                    <span aria-hidden='true' className='absolute inset-0'></span>
                    {product.name}
                  </h3>
                  <p className='mt-1 text-sm text-gray-500'>{product.categories.join(', ')}</p>
                </div>
                <p className='text-sm font-medium text-gray-900'>{formatMoney(product.priceUsd)}</p>
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  )
}
