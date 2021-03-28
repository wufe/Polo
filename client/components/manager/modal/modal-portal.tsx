import { store } from '@/state/models';
import { IModal } from '@/state/models/modal-model';
import { observer } from 'mobx-react-lite';
import React from 'react';
import { createPortal } from 'react-dom';
import './modal.scss';

export const ModalContainer = observer((props: React.PropsWithChildren<{}>) => {

    const disposeModal = () =>
        store.app.modal.setVisible(false);

    return <div className="modal" onClick={() => disposeModal()}>
        <div className="__content" onClick={e => e.stopPropagation()}>
            {props.children}
        </div>
    </div>;
})

export const ModalPortal = observer((props: React.PropsWithChildren<{ modal: IModal, name: string }>) => {

    return (props.modal.visible &&
        props.modal.name === props.name &&
        createPortal(<ModalContainer>
            {props.children}
        </ModalContainer>, document.getElementById('modal'))) || null;
});

export const Modal = (props: React.PropsWithChildren<{ name: string }>) => {
    return <ModalPortal modal={store.app.modal} name={props.name}>{props.children}</ModalPortal>
}