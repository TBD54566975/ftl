export const logLevelCharacter: { [key: number]: string } = {
  1: 't',
  5: 'd',
  9: 'i',
  13: 'w',
  17: 'e',
}

export const logLevelText: { [key: number]: string } = {
  1: 'Trace',
  5: 'Debug',
  9: 'Info',
  13: 'Warn',
  17: 'Error',
}

export const logLevelColor: { [key: number]: string } = {
  1: 'text-gray-400 dark:text-gray-400',
  5: 'text-blue-400 dark:text-blue-400',
  9: 'text-green-500 dark:text-green-300',
  13: 'text-yellow-400 dark:text-yellow-300',
  17: 'text-red-400 dark:text-red-400',
}

export const logLevelBgColor: { [key: number]: string } = {
  1: 'bg-gray-400 dark:bg-gray-400',
  5: 'bg-blue-400 dark:bg-blue-400',
  9: 'bg-green-500 dark:bg-green-300',
  13: 'bg-yellow-400 dark:bg-yellow-300',
  17: 'bg-red-400 dark:bg-red-400',
}

export const logLevelRingColor: { [key: number]: string } = {
  1: 'ring-gray-400 dark:ring-gray-400',
  5: 'ring-blue-400 dark:ring-blue-400',
  9: 'ring-green-500 dark:ring-green-300',
  13: 'ring-yellow-400 dark:ring-yellow-300',
  17: 'ring-red-400 dark:ring-red-400',
}

export const logLevelBadge: { [key: number]: string } = {
  1: `${logLevelColor[1]} bg-blue-300/10 dark:bg-blue-700/30`,
  5: `${logLevelColor[5]} bg-blue-400/10 dark:bg-blue-800/30`,
  9: `${logLevelColor[9]} bg-green-400/30 dark:bg-green-700/30`,
  13: `${logLevelColor[13]} bg-yellow-400/10 dark:bg-yellow-600/30`,
  17: `${logLevelColor[17]} bg-red-500/10 dark:bg-red-700/30`,
}
