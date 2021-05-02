import React, { createRef, Component } from 'react'
import { SelectableGroup } from 'react-selectable-fast'
import { config, IDFromTree } from '../store';
import LazyImage  from '../components/Lazyloading';

class PhotoList extends Component {
  state = {
    disableFirstRow: false,
    reversed: false,
    showSelectableGroup: true,
  }

  countersRef = createRef()

  getSelectableGroupRef = (ref) => {
    window.selectableGroup = ref
  }

  toggleFirstRow = () => {
    this.setState(state => ({ disableFirstRow: !state.disableFirstRow }))
  }

  toggleOrder = () => {
    this.setState(state => ({ reversed: !state.reversed }))
  }

  toggleSelectableGroup = () => {
    this.setState(state => ({
      showSelectableGroup: !state.showSelectableGroup,
    }))
  }

  handleSelecting = (selectingItems) => {
    this.countersRef.current.handleSelecting(selectingItems)
  }

  handleSelectionFinish = selectedItems => {
    console.log('Handle selection finish', selectedItems.length)
    this.countersRef.current.handleSelectionFinish(selectedItems)
  }

  handleSelectedItemUnmount = (_unmountedItem, selectedItems) => {
    console.log('hadneleSelectedItemUnmount')
    this.countersRef.current.handleSelectionFinish(selectedItems)
  }

  handleSelectionClear() {
    console.log('Cancel selection')
  }

  render() {
    const { items } = this.props
    const { showSelectableGroup } = this.state

    return (
      <div style={{width:"100%"}}>
        <button className="btn primary" type="button" onClick={this.toggleFirstRow}>
          Toggle first row
        </button>
        <button className="btn" type="button" onClick={this.toggleOrder}>
          Toggle order
        </button>
        <button className="btn" type="button" onClick={this.toggleSelectableGroup}>
          Toggle group
        </button>
        {showSelectableGroup && (
          <SelectableGroup
            ref={this.getSelectableGroupRef}
            className="main"
            clickClassName="tick"
            enableDeselect={true}
            tolerance={0}
            deselectOnEsc={true}
            allowClickWithoutSelected={false}
            duringSelection={this.handleSelecting}
            onSelectionClear={this.handleSelectionClear}
            onSelectionFinish={this.handleSelectionFinish}
            onSelectedItemUnmount={this.handleSelectedItemUnmount}
            ignoreList={['.not-selectable']}
          >
            {items.map((el, index) => (
              <figure className="galleryImg" ref={selectableRef}>
                <LazyImage src={config.imageUrl + el.id+"?size=tiny&token="+localStorage.getItem('token')} width="100%" height="100%" alt="thumbnail" />
              </figure>
            ))} 
          </SelectableGroup>
        )}
      </div>
    )
  }
}

export default PhotoList