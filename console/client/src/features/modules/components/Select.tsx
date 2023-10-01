import React from 'react'
import { Listbox } from '@headlessui/react'

type OptionTypes = 'verb' | 'module'
interface Option {key: string, value: string, type?: OptionTypes }
const SelectIcon: React.FC<{ type: OptionTypes}> = ({ type }) => (<span>{type}</span>)

export const  Select: React.FC<{
  data: Option[],
  onChange: (value: string) => void 
}> = ({ data, onChange }) => {
  const [selected, setSelected] = React.useState(data[0])
  const handleChange = (option: Option) => {
    setSelected(option)
    onChange(option.value)
  }
  return (
    <Listbox value={selected} onChange={handleChange}>
      <Listbox.Button>{selected.value}</Listbox.Button>
      <Listbox.Options>
        {data.map((item) => (
          <Listbox.Option
            key={item.value}
            value={item}
          >
            {item.type && <SelectIcon type={item.type}/>}{item.value}
          </Listbox.Option>
        ))}
      </Listbox.Options>
    </Listbox>
  )
}
