import { Button, Container, Stack } from '@mui/material'

export const ModulesPage = () => {
  return (
    <>
      <Container maxWidth='sm' sx={{ m: 4 }}>
        <Stack spacing={2}>
          <Button variant='contained'>Sample MUI Button</Button>
          <Button variant='contained'>Sample MUI Button</Button>
        </Stack>
      </Container>
    </>
  )
}
