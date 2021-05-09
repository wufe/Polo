import { store } from ".";

export const useNotification = () => {
    const { addNotification, deleteNotification } = store.app;

    return {
        notify: addNotification,
        dismiss: deleteNotification
    };
}