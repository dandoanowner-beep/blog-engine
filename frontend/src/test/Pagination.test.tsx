import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Pagination from '../components/Pagination'

describe('Pagination', () => {
  it('renders nothing when total <= perPage', () => {
    const { container } = render(
      <Pagination page={1} total={5} perPage={9} onPageChange={vi.fn()} />
    )
    expect(container.firstChild).toBeNull()
  })

  it('renders page buttons', () => {
    render(<Pagination page={1} total={30} perPage={9} onPageChange={vi.fn()} />)
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('2')).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('4')).toBeInTheDocument()
  })

  it('marks current page with aria-current=page', () => {
    render(<Pagination page={2} total={30} perPage={9} onPageChange={vi.fn()} />)
    const btn = screen.getByText('2')
    expect(btn).toHaveAttribute('aria-current', 'page')
  })

  it('calls onPageChange with correct page', async () => {
    const onPageChange = vi.fn()
    render(<Pagination page={1} total={27} perPage={9} onPageChange={onPageChange} />)
    await userEvent.click(screen.getByText('3'))
    expect(onPageChange).toHaveBeenCalledWith(3)
  })

  it('previous button is disabled on first page', () => {
    render(<Pagination page={1} total={27} perPage={9} onPageChange={vi.fn()} />)
    expect(screen.getByLabelText('Previous page')).toBeDisabled()
  })

  it('next button is disabled on last page', () => {
    render(<Pagination page={3} total={27} perPage={9} onPageChange={vi.fn()} />)
    expect(screen.getByLabelText('Next page')).toBeDisabled()
  })

  it('previous button calls onPageChange(page-1)', async () => {
    const onPageChange = vi.fn()
    render(<Pagination page={2} total={27} perPage={9} onPageChange={onPageChange} />)
    await userEvent.click(screen.getByLabelText('Previous page'))
    expect(onPageChange).toHaveBeenCalledWith(1)
  })

  it('next button calls onPageChange(page+1)', async () => {
    const onPageChange = vi.fn()
    render(<Pagination page={1} total={27} perPage={9} onPageChange={onPageChange} />)
    await userEvent.click(screen.getByLabelText('Next page'))
    expect(onPageChange).toHaveBeenCalledWith(2)
  })
})
