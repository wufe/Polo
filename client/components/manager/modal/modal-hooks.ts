import { store } from "@/state/models";

export const useModal = () => {

    const show = (name: string) => {
        store.app.modal.setVisible(true, name);
    };

    const hide = () => {
        store.app.modal.setVisible(false);
    };

    return { show, hide };
}