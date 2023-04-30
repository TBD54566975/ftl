export const modules = [
  {
    id: '76957fed-8c92-4b03-a7e1-6810198fecc3',
    name: 'Time',
    description: 'A simple time service',
    language: 'go',
    status: 'offline',
    environment: 'Staging',
    statusText: 'Initiated 1m 32s ago',
    verbs: [
      {
        id: 'd669783f-202b-4e57-97cd-163e638d1928',
        name: 'Time',
        description: 'Get the current time'
      }
    ]
  },
  {
    id: 'feebd71b-b56e-41d5-8788-74447dc942d3',
    name: 'Echo',
    description: 'A simple echo service',
    language: 'go',
    status: 'online',
    environment: 'Production',
    statusText: 'Initiated 1m 32s ago',
    verbs: [
      {
        id: '48b68beb-1b6c-4e14-91ff-73863a8ece93',
        name: 'Echo',
        description: 'Echos a given message'
      }
    ]
  },
  {
    id: 'f268a0b0-2f1e-4c14-a441-9cc6ad5e986f',
    name: 'Payments',
    description: 'A simple payments service',
    language: 'go',
    status: 'error',
    environment: 'Production',
    statusText: 'Initiated 1m 32s ago',
    verbs: [
      {
        id: '8875541d-340a-420c-b34f-e5efdd799a94',
        name: 'Create',
        description: 'Create a new payment'
      },
      {
        id: '0d97628d-b6c1-468b-b486-a6703bdd1cc3',
        name: 'Delete',
        description: 'Delete a payment'
      }
    ]
  }
]
