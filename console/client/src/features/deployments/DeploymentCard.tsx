import { useNavigate } from 'react-router-dom'
import { Card } from '../../components/Card'
import { deploymentTextColor } from './deployment.utils'

export const DeploymentCard = ({ name, className }: { name: string; className?: string }) => {
  const navigate = useNavigate()
  return (
    <Card key={name} topBarColor='bg-green-500' className={className} onClick={() => navigate(`/deployments/${name}`)}>
      {name}
      <p className={`text-xs ${deploymentTextColor(name)}`}>{name}</p>
    </Card>
  )
}
