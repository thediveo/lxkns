import { atom } from 'jotai'
import { useReducerAtom } from 'jotai/utils'

export const EXPANDALL = "expandall";
export const COLLAPSEALL = "collapseall";

// In order to be able to always trigger an action, we have to mutate an action
// state (atom).
export interface Action {
    action: string // the action itself
    mutation?: number // for retriggering any action
}

// Our action reducer ensures that the new action always differs from the
// previous action by mutation.
const actionReducer = (prev: Action, action: Action): Action => ({
    action: action.action,
    mutation: (prev.mutation + 1) & 0x0f // who'll ever need more than 4 bits?!
})

export type ActionSetter = (newaction: string) => void
export interface ActionUsage extends Array<Action | ActionSetter> { 0: Action, 1: ActionSetter }

const treeActionAtom = atom({ action: "", mutation: 0 } as Action)

export const useTreeAction = (): ActionUsage => {
    const [action, dispatch] = useReducerAtom(treeActionAtom, actionReducer)
    return [action, (newaction: string) => dispatch({ action: newaction })]
}
