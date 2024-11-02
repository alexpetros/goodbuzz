class EditableField extends HTMLElement {
  connectedCallback() {
    this.isEditing = false
    this.name = this.getAttribute('name')
    this.value = this.getAttribute('value') || ''
    this._internals = this.attachInternals()

    // I want to call this onsave but I can't because of stupid templ
    this.onSave = new Function(this.getAttribute('savefunc'))
    this.onSave.bind(this)

    this.clickListener = this.addEventListener('click', (e) => {
      if (!this.isEditing && e.detail > 1) {
        this.switchToEdit()
      }
    })

    this.keydownListener = this.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        const input = this.querySelector('input')
        this.value = input?.value || ''

        this.switchToView()
        this.onSave()
      }
    })

    // Start off in the non-edit view
    this.switchToView()
  }

  // Close our the event listeners when the editable leaves
  disconnectedCallback() {
    this.removeEventListener('click', this.clickListener)
    this.removeEventListener('keydown', this.keydownListener)
  }

  switchToView() {
    this._internals.states.delete('editing')

    this.innerHTML = `<span class="value">${this.value}</span>
    <input type="hidden" name="${this.name}" value="${this.value}">
    `
  }

  switchToEdit() {
    this._internals.states.add('editing')
    this.innerHTML = `<input class="input" value="${this.value}">`
  }
}

customElements.define('editable-field', EditableField)

