import { Instance, types } from "mobx-state-tree";

export const ModalModel = types.model({
    visible: types.optional(types.boolean, false),
    name   : types.optional(types.string, ''),
}).actions(self => {

    const setVisible = (visible: boolean, name?: string) => {
        self.visible = visible;
        self.name = name ?? '';
    };

    return { setVisible };
})

export interface IModal extends Instance<typeof ModalModel> {}

export const initialModalState = ModalModel.create({});