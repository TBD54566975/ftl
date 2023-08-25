import {createContext, useState} from 'react';

interface SidePanelContextType {
  isOpen: boolean;
  component: React.ReactNode;
  openPanel: (component: React.ReactNode) => void;
  closePanel: () => void;
}

const defaultContextValue: SidePanelContextType = {
  isOpen: false,
  component: null,
  openPanel: () => {},
  closePanel: () => {},
};

export const SidePanelContext =
  createContext<SidePanelContextType>(defaultContextValue);

export const SidePanelProvider = ({children}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [component, setComponent] = useState(null);

  const openPanel = comp => {
    setIsOpen(true);
    setComponent(comp);
  };

  const closePanel = () => {
    setIsOpen(false);
    setComponent(null);
  };

  return (
    <SidePanelContext.Provider
      value={{isOpen, openPanel, closePanel, component}}>
      {children}
    </SidePanelContext.Provider>
  );
};
