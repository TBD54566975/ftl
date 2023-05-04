import { classNames } from '../../utils'
import { statuses } from '../../data/Types'
import { useSchema } from '../../hooks/use-schema'
import { Card } from '../../components/Card'

export default function ModulesPage() {
  const schema = useSchema()

  return (
    <>
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        {schema.map(module => (
          <Card key={module.schema?.name}>
            <div className="min-w-0 flex-1">
              <a
                href={`modules/${module.schema?.name}`}
                className="focus:outline-none"
              >
                <span className="absolute inset-0" aria-hidden="true" />
                <div className="min-w-0 flex-auto">
                  <div className="flex items-center gap-x-3">
                    <div
                      className={classNames(
                        statuses['online'],
                        'flex-none rounded-full p-1'
                      )}
                    >
                      <div className="h-2 w-2 rounded-full bg-current" />
                    </div>
                    <p className="text-sm font-medium text-gray-900 dark:text-gray-300">
                      {module.schema?.name}
                    </p>
                  </div>
                </div>

                {(module.schema?.comments.length ?? 0) > 0 && (
                  <div className="min-w-0 flex-auto pt-2">
                    <div className="flex items-center gap-x-3">
                      <p className="truncate text-sm text-gray-500">
                        {module.schema?.comments}
                      </p>
                    </div>
                  </div>
                )}
              </a>
            </div>
          </Card>
        ))}
      </div>
    </>
  )
}
